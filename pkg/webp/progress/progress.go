package progress

import (
	"h2blog/internal/model/dto"
	"sync"
	"sync/atomic"
)

// Tracker 进度追踪器接口
type Tracker interface {
	Subscribe(clientID string) chan dto.ImgDto
	Unsubscribe(clientID string)
	UpdateProgress(imgDto dto.ImgDto, success bool)
	GetProgress() (total, success, failed int32)
}

// ProgressTracker 进度追踪器实现
type ProgressTracker struct {
	Total     int32
	Success   int32
	Failed    int32
	mu        sync.RWMutex
	observers map[string]chan dto.ImgDto
}

func NewTracker(total int32) Tracker {
	return &ProgressTracker{
		Total:     total,
		observers: make(map[string]chan dto.ImgDto),
	}
}

// Subscribe 订阅进度更新
func (p *ProgressTracker) Subscribe(clientID string) chan dto.ImgDto {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan dto.ImgDto, 10)
	p.observers[clientID] = ch
	return ch
}

// Unsubscribe 取消订阅
func (p *ProgressTracker) Unsubscribe(clientID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ch, exists := p.observers[clientID]; exists {
		close(ch)
		delete(p.observers, clientID)
	}
}

// UpdateProgress 更新进度
func (p *ProgressTracker) UpdateProgress(imgDto dto.ImgDto, success bool) {
	if success {
		atomic.AddInt32(&p.Success, 1)
	} else {
		atomic.AddInt32(&p.Failed, 1)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, ch := range p.observers {
		select {
		case ch <- imgDto:
		default:
		}
	}
}

// GetProgress 获取当前进度
func (p *ProgressTracker) GetProgress() (total, success, failed int32) {
	return p.Total, atomic.LoadInt32(&p.Success), atomic.LoadInt32(&p.Failed)
}
