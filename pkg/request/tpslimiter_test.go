package request

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	a := assert.New(t)
	l := NewTPSLimiter()
	finished := make(chan struct{})
	go func() {
		l.Limit(context.Background(), "token", 1, 1)
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		a.Fail("Limit should be finished instantly.")
	}

	finished = make(chan struct{})
	go func() {
		l.Limit(context.Background(), "token", 1, 1)
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		a.Fail("Limit should be finished in 1 second.")
	}

	finished = make(chan struct{})
	go func() {
		l.Limit(context.Background(), "token", 10, 1)
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(1 * time.Second):
		a.Fail("Limit should be finished instantly.")
	}

}
