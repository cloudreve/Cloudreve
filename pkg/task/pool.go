package task

import "sync"

// Pool 带有最大配额的goroutines任务池
type Pool struct {
	// 容量
	capacity int
	// 初始容量
	initialCapacity int

	// 终止信号
	terminateSignal chan error
	// 全部任务完成的信号
	finishSignal chan bool
	// 有空余位置的信号
	freeSignal chan bool

	// 是否已关闭
	closed bool
	// 是否正在等待任务结束
	waiting bool

	// 互斥锁
	lock sync.Mutex
	// 等待队列
	pending []Job
}

// Job 任务
type Job interface {
	// 任务处理方法，如果error不为nil，
	// 任务池会关闭并中止接受新任务
	Do() error
}

// NewGoroutinePool 创建一个容量为capacity的任务池
func NewGoroutinePool(capacity int) *Pool {
	pool := &Pool{
		capacity:        capacity,
		initialCapacity: capacity,
		terminateSignal: make(chan error),
		finishSignal:    make(chan bool),
		freeSignal:      make(chan bool),
	}
	go pool.Schedule()
	return pool
}

// Schedule 为等待队列的任务分配Worker，以及检测错误状态、所有任务完成
func (pool *Pool) Schedule() {
	for {
		select {
		case <-pool.freeSignal:
			// 有新的空余名额
			pool.lock.Lock()
			if len(pool.pending) > 0 {
				// 有待处理的任务，开始处理
				var job Job
				job, pool.pending = pool.pending[0], pool.pending[1:]
				go pool.start(job)
			} else {
				if pool.waiting && pool.capacity == pool.initialCapacity {
					// 所有任务已结束
					pool.lock.Unlock()
					pool.finishSignal <- true
					return
				}
				pool.lock.Unlock()
			}
		case <-pool.terminateSignal:
			// 有任务意外中止，则发送完成信号
			pool.finishSignal <- true
			return
		}
	}
}

// Wait 等待队列中所有任务完成或有Job返回错误中止
func (pool *Pool) Wait() {
	pool.lock.Lock()
	pool.waiting = true
	pool.lock.Unlock()
	_ = <-pool.finishSignal
}

// Submit 提交新任务
func (pool *Pool) Submit(job Job) {
	if pool.closed {
		return
	}

	pool.lock.Lock()
	if pool.capacity < 1 {
		// 容量为空时，加入等待队列
		pool.pending = append(pool.pending, job)
		pool.lock.Unlock()
		return
	}

	// 还有空闲容量时，开始执行任务
	go pool.start(job)
}

// 开始执行任务
func (pool *Pool) start(job Job) {
	pool.capacity--
	pool.lock.Unlock()

	err := job.Do()
	if err != nil {
		pool.closed = true
		close(pool.terminateSignal)
	}

	pool.lock.Lock()
	pool.capacity++
	pool.lock.Unlock()
	pool.freeSignal <- true
}
