package task

type Pool struct {
	// 容量
	capacity int
	// 终止信号
	terminateSignal chan error
	// 全部任务完成的信号
	finishSignal chan bool
}

type Worker interface {
	Do() error
}

func (pool *Pool) Submit(worker Worker) {
	err := worker.Do()
	if err != nil {
		close(pool.terminateSignal)
	}
}
