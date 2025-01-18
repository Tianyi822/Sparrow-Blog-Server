package webp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"h2blog/internal/model/dto"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/pkg/utils"
	"h2blog/storage"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

var (
	Converter     *converter // 全局转换器实例
	converterOnce sync.Once  // 确保转换器只初始化一次
)

// task 定义单个转换任务
// 包含任务上下文和图片信息
type task struct {
	ctx    context.Context // 任务上下文
	imgDto dto.ImgDto      // 图片信息
}

// OutputData 定义输出数据结构
type OutputData struct {
	ImgDto dto.ImgDto // 转换后的图片信息
	Flag   bool       // 转换是否成功
}

// CompletionStatus 完成状态
type CompletionStatus struct {
	Success bool      // 是否成功
	Message string    // 状态消息
	Time    time.Time // 完成时间
}

// converter WebP转换器
// 负责管理WebP格式转换任务队列和工作协程
type converter struct {
	inputCh   chan task             // 输入任务通道
	outputCh  chan OutputData       // 输出结果通道
	quality   float32               // WebP转换质量
	wg        sync.WaitGroup        // 等待组，用于等待所有工作协程完成
	done      chan struct{}         // 关闭信号通道
	completed chan CompletionStatus // 完成状态通知通道
	workerNum int                   // 工作协程数量
	timeout   time.Duration         // 任务超时时间
	taskCount atomic.Int32          // 添加任务计数器
}

// InitConverter 初始化WebP转换器
// 创建转换器实例并启动工作协程
func InitConverter() {
	converterOnce.Do(func() {
		// 创建转换器实例
		Converter = &converter{
			inputCh:   make(chan task, 30),            // 输入任务通道，缓冲区大小为30
			outputCh:  make(chan OutputData, 30),      // 输出结果通道，缓冲区大小为30
			quality:   config.UserConfig.WebP.Quality, // WebP转换质量
			done:      make(chan struct{}),            // 关闭信号通道
			completed: make(chan CompletionStatus, 1), // 完成状态通道，缓冲区大小为1
			workerNum: runtime.NumCPU() / 2,           // 工作协程数量，等于CPU核心数除以2
			timeout:   5 * time.Minute,                // 任务处理超时时间
		}

		// 启动工作协程
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
	// 设置总任务数
	c.taskCount.Store(int32(len(dtos)))

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

// GetOutputCh 获取输入通道
func (c *converter) GetOutputCh() chan OutputData {
	return c.outputCh
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
		case task, ok := <-c.inputCh:
			if !ok {
				return
			}

			// 处理接收到的任务
			c.handleTask(task)

			// 任务完成后
			remaining := c.taskCount.Add(-1)
			if remaining == 0 {
				c.NotifyCompletion(true, "任务全部完成")
			}
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
	if err != nil {
		logger.Error("处理任务失败: %v", err)
	}
	// 发送处理结果到输出通道
	c.outputCh <- OutputData{
		ImgDto: task.imgDto,
		Flag:   err == nil,
	}

	// 立即释放上下文资源
	cancel()
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
		msg := fmt.Sprintf("下载图片失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 将图片转换为WebP格式
	converted, err := convertToWebP(imgBytes, c.quality)
	if err != nil {
		msg := fmt.Sprintf("转换图片失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 将转换后的WebP图片上传到OSS
	if err := storage.Storage.PutContentToOss(ctx, converted, utils.GenOssSavePath(task.imgDto.ImgName, utils.Webp)); err != nil {
		msg := fmt.Sprintf("上传图片失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
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
	close(c.inputCh)
	close(c.outputCh)
	close(c.done)
	c.wg.Wait()
}

func (c *converter) IsEmpty() bool {
	return c.taskCount.Load() == 0
}

// NotifyCompletion 发送完成通知
func (c *converter) NotifyCompletion(success bool, msg string) {
	c.completed <- CompletionStatus{
		Success: success,
		Message: msg,
		Time:    time.Now(),
	}
}

// GetCompletionStatus 获取完成状态
func (c *converter) GetCompletionStatus() <-chan CompletionStatus {
	return c.completed
}
