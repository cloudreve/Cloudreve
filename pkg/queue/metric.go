package queue

import "sync/atomic"

// Metric interface
type Metric interface {
	IncBusyWorker()
	DecBusyWorker()
	BusyWorkers() uint64
	SuccessTasks() uint64
	FailureTasks() uint64
	SubmittedTasks() uint64
	IncSuccessTask()
	IncFailureTask()
	IncSubmittedTask()
}

var _ Metric = (*metric)(nil)

type metric struct {
	busyWorkers     uint64
	successTasks    uint64
	failureTasks    uint64
	submittedTasks  uint64
	suspendingTasks uint64
}

// NewMetric for default metric structure
func NewMetric() Metric {
	return &metric{}
}

func (m *metric) IncBusyWorker() {
	atomic.AddUint64(&m.busyWorkers, 1)
}

func (m *metric) DecBusyWorker() {
	atomic.AddUint64(&m.busyWorkers, ^uint64(0))
}

func (m *metric) BusyWorkers() uint64 {
	return atomic.LoadUint64(&m.busyWorkers)
}

func (m *metric) IncSuccessTask() {
	atomic.AddUint64(&m.successTasks, 1)
}

func (m *metric) IncFailureTask() {
	atomic.AddUint64(&m.failureTasks, 1)
}

func (m *metric) IncSubmittedTask() {
	atomic.AddUint64(&m.submittedTasks, 1)
}

func (m *metric) SuccessTasks() uint64 {
	return atomic.LoadUint64(&m.successTasks)
}

func (m *metric) FailureTasks() uint64 {
	return atomic.LoadUint64(&m.failureTasks)
}

func (m *metric) SubmittedTasks() uint64 {
	return atomic.LoadUint64(&m.submittedTasks)
}

func (m *metric) SuspendingTasks() uint64 {
	return atomic.LoadUint64(&m.suspendingTasks)
}

func (m *metric) IncSuspendingTask() {
	atomic.AddUint64(&m.suspendingTasks, 1)
}

func (m *metric) DecSuspendingTask() {
	atomic.AddUint64(&m.suspendingTasks, ^uint64(0))
}
