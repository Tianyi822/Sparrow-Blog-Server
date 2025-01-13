package webp

import (
	"h2blog/internal/model/dto"
	"sync"
	"sync/atomic"
)

// Tracker 进度追踪器实现
type Tracker struct {
	Total     int32                      // 总任务数
	Success   int32                      // 成功任务数
	Failed    int32                      // 失败任务数
	mu        sync.RWMutex               // 读写锁，用于保护observers
	observers map[string]chan dto.ImgDto // 观察者列表，key为客户端ID，value为通知channel
}

// NewTracker 创建新的进度追踪器
// total: 总任务数
// 返回值: 新的Tracker实例
func NewTracker(total int32) *Tracker {
	return &Tracker{
		Total:     total,
		observers: make(map[string]chan dto.ImgDto),
	}
}

// Subscribe 订阅进度更新
// clientID: 客户端唯一标识
// 返回值: 用于接收进度更新的channel
func (p *Tracker) Subscribe(clientID string) chan dto.ImgDto {
	// 加锁保证线程安全
	p.mu.Lock()
	// 确保锁在函数返回时释放
	defer p.mu.Unlock()

	// 创建带缓冲的channel，容量为10
	ch := make(chan dto.ImgDto, 10)
	// 将客户端ID与channel关联
	p.observers[clientID] = ch
	return ch
}

// Unsubscribe 取消订阅
// clientID: 要取消订阅的客户端ID
func (p *Tracker) Unsubscribe(clientID string) {
	// 加锁保证线程安全
	p.mu.Lock()
	// 确保锁在函数返回时释放
	defer p.mu.Unlock()

	// 检查客户端是否存在
	if ch, exists := p.observers[clientID]; exists {
		// 关闭channel防止goroutine泄漏
		close(ch)
		// 从观察者列表中移除
		delete(p.observers, clientID)
	}
}

// UpdateProgress 更新处理进度
// imgDto: 图片处理结果
// success: 是否处理成功
func (p *Tracker) UpdateProgress(imgDto dto.ImgDto, success bool) {
	// 使用原子操作更新成功/失败计数
	if success {
		atomic.AddInt32(&p.Success, 1) // 成功计数+1
	} else {
		atomic.AddInt32(&p.Failed, 1) // 失败计数+1
	}

	// 记录需要移除的无响应观察者
	var deadObservers []string

	// 使用读锁保护观察者列表
	p.mu.RLock()
	// 遍历所有观察者
	for id, ch := range p.observers {
		select {
		case ch <- imgDto: // 尝试发送进度更新
			// 成功发送到观察者
		default:
			// channel已满或关闭，标记为需要移除
			deadObservers = append(deadObservers, id)
		}
	}
	p.mu.RUnlock()

	// 如果有需要移除的观察者
	if len(deadObservers) > 0 {
		// 使用写锁修改观察者列表
		p.mu.Lock()
		// 遍历需要移除的观察者
		for _, id := range deadObservers {
			if ch, exists := p.observers[id]; exists {
				close(ch)               // 关闭channel
				delete(p.observers, id) // 从map中移除
			}
		}
		p.mu.Unlock()
	}
}

// GetProgress 获取当前进度
// 返回值:
//
//	total: 总任务数
//	success: 成功数
//	failed: 失败数
func (p *Tracker) GetProgress() (total, success, failed int32) {
	// 使用原子操作读取当前进度
	return p.Total,
		atomic.LoadInt32(&p.Success), // 读取成功计数
		atomic.LoadInt32(&p.Failed) // 读取失败计数
}
