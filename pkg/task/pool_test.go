package task

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type testJob struct {
	id    string
	sleep time.Duration
}

func (job testJob) Do() error {
	fmt.Printf("任务%s开始执行\n", job.id)
	time.Sleep(job.sleep * time.Second)
	fmt.Printf("任务%s执行完毕\n", job.id)
	if job.id == "3" {
		return errors.New("error")
	}
	return nil
}

func TestPool_Submit(t *testing.T) {
	pool := NewGoroutinePool(1)
	task1 := testJob{
		id:    "1",
		sleep: 5,
	}
	task2 := testJob{
		id:    "2",
		sleep: 5,
	}
	task3 := testJob{
		id:    "3",
		sleep: 2,
	}
	task4 := testJob{
		id:    "4",
		sleep: 5,
	}
	pool.Submit(task1)
	pool.Submit(task2)
	pool.Submit(task3)
	pool.Submit(task4)
	pool.Wait()
}
