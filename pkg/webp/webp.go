package webp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"h2blog/pkg/config"
	"h2blog/storage"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"sync"
	"time"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

// TaskProgress 任务进度跟踪
type TaskProgress struct {
	Total     int
	Completed int
	Failed    int
	mu        sync.Mutex
}

// Task 定义转换任务
type task struct {
	ctx      context.Context
	ossPath  string
	progress *TaskProgress
}

// Converter 结构体用于处理数据转换任务
type converter struct {
	inputCh   chan task
	quality   float32
	wg        sync.WaitGroup
	done      chan struct{}
	workerNum int
	timeout   time.Duration
}

var (
	Converter     *converter
	converterOnce sync.Once
)

// InitConverter 初始化转换器
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

// AddBatchTasks 批量添加任务
func (c *converter) AddBatchTasks(ctx context.Context, paths []string) (*TaskProgress, error) {
	progress := &TaskProgress{
		Total: len(paths),
	}

	for _, path := range paths {
		if err := c.AddTaskWithProgress(ctx, path, progress); err != nil {
			return progress, err
		}
	}
	return progress, nil
}

// AddTaskWithProgress 添加带进度的任务
func (c *converter) AddTaskWithProgress(ctx context.Context, ossPath string, progress *TaskProgress) error {
	return c.AddTaskWithCallback(ctx, ossPath, progress)
}

// AddTaskWithCallback 添加带回调的任务
func (c *converter) AddTaskWithCallback(ctx context.Context, ossPath string, progress *TaskProgress) error {
	select {
	case c.inputCh <- task{
		ctx:      ctx,
		ossPath:  ossPath,
		progress: progress,
	}:
		return nil
	default:
		return errors.New("转换队列已满")
	}
}

// UpdateProgress 更新进度
func (p *TaskProgress) UpdateProgress(success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if success {
		p.Completed++
	} else {
		p.Failed++
	}
}

// GetProgress 获取进度
func (p *TaskProgress) GetProgress() (completed, failed, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Completed, p.Failed, p.Total
}

// startWorker 启动工作协程
func (c *converter) startWorker() {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case task := <-c.inputCh:
			ctx, cancel := context.WithTimeout(task.ctx, c.timeout)
			err := c.processTask(ctx, task)

			if task.progress != nil {
				task.progress.UpdateProgress(err == nil)
			}

			cancel()
		case <-c.done:
			return
		}
	}
}

// processTask 处理任务
func (c *converter) processTask(ctx context.Context, task task) error {
	imgBytes, err := storage.Storage.GetContentFromOss(ctx, task.ossPath)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}

	converted, err := convertToWebP(imgBytes, c.quality)
	if err != nil {
		return fmt.Errorf("转换WebP失败: %w", err)
	}

	newPath := strings.Split(task.ossPath, ".")
	webpPath := newPath[0] + ".webp"
	if err := storage.Storage.PutContentToOss(ctx, converted, webpPath); err != nil {
		return fmt.Errorf("上传WebP失败: %w", err)
	}

	return nil
}

// convertToWebP 转换为WebP格式
func convertToWebP(src []byte, quality float32) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, quality)
	if err != nil {
		return nil, fmt.Errorf("创建编码器选项失败: %w", err)
	}

	if err := webp.Encode(&buf, img, options); err != nil {
		return nil, fmt.Errorf("编码WebP失败: %w", err)
	}

	return buf.Bytes(), nil
}

// Shutdown 优雅关闭转换器
func (c *converter) Shutdown() {
	close(c.done)
	c.wg.Wait()
}
