package queue

import (
	"errors"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrQueueShutdown the queue is released and closed.
	ErrQueueShutdown = errors.New("queue has been closed and released")
	// ErrMaxCapacity Maximum size limit reached
	ErrMaxCapacity = errors.New("golang-queue: maximum size limit reached")
	// ErrNoTaskInQueue there is nothing in the queue
	ErrNoTaskInQueue = errors.New("golang-queue: no Task in queue")
)

type (
	Scheduler interface {
		// Queue add a new Task into the queue
		Queue(task Task) error
		// Request get a new Task from the queue
		Request() (Task, error)
		// Shutdown stop all worker
		Shutdown() error
	}
	fifoScheduler struct {
		sync.Mutex
		taskQueue taskHeap
		capacity  int
		count     int
		exit      chan struct{}
		logger    logging.Logger
		stopOnce  sync.Once
		stopFlag  int32
	}
	taskHeap []Task
)

// Queue send Task to the buffer channel
func (s *fifoScheduler) Queue(task Task) error {
	if atomic.LoadInt32(&s.stopFlag) == 1 {
		return ErrQueueShutdown
	}
	if s.capacity > 0 && s.count >= s.capacity {
		return ErrMaxCapacity
	}

	s.Lock()
	s.taskQueue.Push(task)
	s.count++
	s.Unlock()

	return nil
}

// Request a new Task from channel
func (s *fifoScheduler) Request() (Task, error) {
	if atomic.LoadInt32(&s.stopFlag) == 1 {
		return nil, ErrQueueShutdown
	}

	if s.count == 0 {
		return nil, ErrNoTaskInQueue
	}
	s.Lock()
	if s.taskQueue[s.taskQueue.Len()-1].ResumeTime() > time.Now().Unix() {
		s.Unlock()
		return nil, ErrNoTaskInQueue
	}

	data := s.taskQueue.Pop()
	s.count--
	s.Unlock()

	return data.(Task), nil
}

// Shutdown the worker
func (s *fifoScheduler) Shutdown() error {
	if !atomic.CompareAndSwapInt32(&s.stopFlag, 0, 1) {
		return ErrQueueShutdown
	}

	return nil
}

// NewFifoScheduler for create new Scheduler instance
func NewFifoScheduler(queueSize int, logger logging.Logger) Scheduler {
	w := &fifoScheduler{
		taskQueue: make([]Task, 2),
		capacity:  queueSize,
		logger:    logger,
	}

	return w
}

// Implement heap.Interface
func (h taskHeap) Len() int {
	return len(h)
}

func (h taskHeap) Less(i, j int) bool {
	return h[i].ResumeTime() < h[j].ResumeTime()
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *taskHeap) Push(x any) {
	*h = append(*h, x.(Task))
}

func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
