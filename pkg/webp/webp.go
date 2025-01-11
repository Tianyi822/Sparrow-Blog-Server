package webp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"h2blog/internal/model/dto"
	"h2blog/pkg/config"
	"h2blog/pkg/utils"
	"h2blog/pkg/webp/progress"
	"h2blog/storage"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"sync"
	"time"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

var (
	Converter     *converter
	converterOnce sync.Once
)

// task 定义单个转换任务
// 包含任务上下文和图片信息
type task struct {
	ctx    context.Context
	imgDto dto.ImgDto
}

// converter WebP转换器
// 负责管理WebP格式转换任务队列和工作协程
type converter struct {
	inputCh   chan task
	quality   float32
	wg        sync.WaitGroup
	done      chan struct{}
	workerNum int
	timeout   time.Duration
	progress  progress.Tracker
}

// InitConverter 初始化WebP转换器
// 创建转换器实例并启动工作协程
func InitConverter() {
	converterOnce.Do(func() {
		Converter = &converter{
			inputCh:   make(chan task, 30),
			quality:   config.UserConfig.WebP.Quality,
			done:      make(chan struct{}),
			workerNum: 3,
			timeout:   5 * time.Minute,
		}

		for i := 0; i < Converter.workerNum; i++ {
			go Converter.startWorker()
		}
	})
}

// AddBatchTasks 批量添加转换任务
// 将多个图片转换任务加入处理队列
// 参数：
//   - ctx: 上下文
//   - dtos: 图片信息列表
//
// 返回：
//   - error: 添加失败时返回错误
func (c *converter) AddBatchTasks(ctx context.Context, dtos []dto.ImgDto) error {
	// 创建进度追踪器
	c.progress = progress.NewTracker(int32(len(dtos)))

	for _, imgDto := range dtos {
		select {
		case c.inputCh <- task{
			ctx:    ctx,
			imgDto: imgDto,
		}:
			continue
		default:
			return errors.New("转换队列已满")
		}
	}
	return nil
}

// startWorker 启动工作协程
// 每个工作协程从任务队列中获取任务并处理
func (c *converter) startWorker() {
	// 增加等待组计数
	c.wg.Add(1)
	// 确保工作结束时减少计数
	defer c.wg.Done()

	// 持续处理任务
	for {
		select {
		case task := <-c.inputCh:
			// 处理接收到的任务
			c.handleTask(task)
		case <-c.done:
			// 收到关闭信号，退出协程
			return
		}
	}
}

func (c *converter) handleTask(task task) {
	// 创建带超时的上下文，防止任务卡住
	ctx, cancel := context.WithTimeout(task.ctx, c.timeout)

	// 执行实际的任务处理
	err := c.processTask(ctx, task)

	// 立即释放上下文资源
	cancel()

	// 更新进度
	if c.progress != nil {
		c.progress.UpdateProgress(task.imgDto, err == nil)
	}
}

// GetProgress 获取进度追踪器
func (c *converter) GetProgress() progress.Tracker {
	return c.progress
}

// processTask 处理单个转换任务
// 执行图片下载、格式转换和上传操作
// 参数：
//   - ctx: 上下文
//   - task: 转换任务
//
// 返回：
//   - error: 处理过程中发生的错误
func (c *converter) processTask(ctx context.Context, task task) error {
	// 从OSS下载原始图片
	imgBytes, err := storage.Storage.GetContentFromOss(ctx, utils.GenOssSavePath(task.imgDto.ImgName, task.imgDto.ImgType))
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}

	// 将图片转换为WebP格式
	converted, err := convertToWebP(imgBytes, c.quality)
	if err != nil {
		return fmt.Errorf("转换WebP失败: %w", err)
	}

	// 将转换后的WebP图片上传到OSS
	if err := storage.Storage.PutContentToOss(ctx, converted, utils.GenOssSavePath(task.imgDto.ImgName, utils.Webp)); err != nil {
		return fmt.Errorf("上传WebP失败: %w", err)
	}

	return nil
}

// convertToWebP 将图片转换为WebP格式
// 参数：
//   - src: 原始图片字节数据
//   - quality: WebP质量参数（0-100）
//
// 返回：
//   - []byte: 转换后的WebP字节数据
//   - error: 转换过程中发生的错误
func convertToWebP(src []byte, quality float32) ([]byte, error) {
	// 解码原始图片数据
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 创建缓冲区存储转换结果
	var buf bytes.Buffer

	// 创建WebP编码选项
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, quality)
	if err != nil {
		return nil, fmt.Errorf("创建编码器选项失败: %w", err)
	}

	// 执行WebP编码
	if err := webp.Encode(&buf, img, options); err != nil {
		return nil, fmt.Errorf("编码WebP失败: %w", err)
	}

	// 返回转换后的WebP数据
	return buf.Bytes(), nil
}

// Shutdown 优雅关闭转换器
// 关闭任务通道并等待所有工作协程退出
func (c *converter) Shutdown() {
	close(c.done)
	c.wg.Wait()
}
