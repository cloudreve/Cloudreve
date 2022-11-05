package backoff

import (
	"errors"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"net/http"
	"strconv"
	"time"
)

// Backoff used for retry sleep backoff
type Backoff interface {
	Next(err error) bool
	Reset()
}

// ConstantBackoff implements Backoff interface with constant sleep time. If the error
// is retryable and with `RetryAfter` defined, the `RetryAfter` will be used as sleep duration.
type ConstantBackoff struct {
	Sleep time.Duration
	Max   int

	tried int
}

func (c *ConstantBackoff) Next(err error) bool {
	c.tried++
	if c.tried > c.Max {
		return false
	}

	var e *RetryableError
	if errors.As(err, &e) && e.RetryAfter > 0 {
		util.Log().Warning("Retryable error %q occurs in backoff, will sleep after %s.", e, e.RetryAfter)
		time.Sleep(e.RetryAfter)
	} else {
		time.Sleep(c.Sleep)
	}

	return true
}

func (c *ConstantBackoff) Reset() {
	c.tried = 0
}

type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

// NewRetryableErrorFromHeader constructs a new RetryableError from http response header
// and existing error.
func NewRetryableErrorFromHeader(err error, header http.Header) *RetryableError {
	retryAfter := header.Get("retry-after")
	if retryAfter == "" {
		retryAfter = "0"
	}

	res := &RetryableError{
		Err: err,
	}

	if retryAfterSecond, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
		res.RetryAfter = time.Duration(retryAfterSecond) * time.Second
	}

	return res
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error with retry-after=%s: %s", e.RetryAfter, e.Err)
}
