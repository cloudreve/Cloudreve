package backoff

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConstantBackoff_Next(t *testing.T) {
	a := assert.New(t)

	b := &ConstantBackoff{Sleep: time.Duration(0), Max: 3}
	a.True(b.Next())
	a.True(b.Next())
	a.True(b.Next())
	a.False(b.Next())
	b.Reset()
	a.True(b.Next())
	a.True(b.Next())
	a.True(b.Next())
	a.False(b.Next())
}
