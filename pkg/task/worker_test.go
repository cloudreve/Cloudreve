package task

import (
	"testing"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/stretchr/testify/assert"
)

type MockJob struct {
	Err    *JobError
	Status int
	DoFunc func()
}

func (job *MockJob) Type() int {
	panic("implement me")
}

func (job *MockJob) Creator() uint {
	panic("implement me")
}

func (job *MockJob) Props() string {
	panic("implement me")
}

func (job *MockJob) Model() *model.Task {
	panic("implement me")
}

func (job *MockJob) SetStatus(status int) {
	job.Status = status
}

func (job *MockJob) Do() {
	job.DoFunc()
}

func (job *MockJob) SetError(*JobError) {
}

func (job *MockJob) GetError() *JobError {
	return job.Err
}

func TestGeneralWorker_Do(t *testing.T) {
	asserts := assert.New(t)
	worker := &GeneralWorker{}
	job := &MockJob{}

	// 正常
	{
		job.DoFunc = func() {
		}
		worker.Do(job)
		asserts.Equal(Complete, job.Status)
	}

	// 有错误
	{
		job.DoFunc = func() {
		}
		job.Status = Queued
		job.Err = &JobError{Msg: "error"}
		worker.Do(job)
		asserts.Equal(Error, job.Status)
	}

	// 有致命错误
	{
		job.DoFunc = func() {
			panic("mock fatal error")
		}
		job.Status = Queued
		job.Err = nil
		worker.Do(job)
		asserts.Equal(Error, job.Status)
	}

}
