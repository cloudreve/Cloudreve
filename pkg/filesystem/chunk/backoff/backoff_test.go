package backoff

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestConstantBackoff_Next(t *testing.T) {
	a := assert.New(t)

	// General error
	{
		err := errors.New("error")
		b := &ConstantBackoff{Sleep: time.Duration(0), Max: 3}
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.False(b.Next(err))
		b.Reset()
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.False(b.Next(err))
	}

	// Retryable error
	{
		err := &RetryableError{RetryAfter: time.Duration(1)}
		b := &ConstantBackoff{Sleep: time.Duration(0), Max: 3}
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.False(b.Next(err))
		b.Reset()
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.True(b.Next(err))
		a.False(b.Next(err))
	}

}

func TestNewRetryableErrorFromHeader(t *testing.T) {
	a := assert.New(t)
	// no retry-after header
	{
		err := NewRetryableErrorFromHeader(nil, http.Header{})
		a.Empty(err.RetryAfter)
	}

	// with retry-after header
	{
		header := http.Header{}
		header.Add("retry-after", "120")
		err := NewRetryableErrorFromHeader(nil, header)
		a.EqualValues(time.Duration(120)*time.Second, err.RetryAfter)
	}
}
