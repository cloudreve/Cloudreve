package backoff

import "time"

// Backoff used for retry sleep backoff
type Backoff interface {
	Next() bool
	Reset()
}

// ConstantBackoff implements Backoff interface with constant sleep time
type ConstantBackoff struct {
	Sleep time.Duration
	Max   int

	tried int
}

func (c *ConstantBackoff) Next() bool {
	c.tried++
	if c.tried > c.Max {
		return false
	}

	time.Sleep(c.Sleep)
	return true
}

func (c *ConstantBackoff) Reset() {
	c.tried = 0
}
