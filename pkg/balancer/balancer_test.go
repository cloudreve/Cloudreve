package balancer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBalancer(t *testing.T) {
	a := assert.New(t)
	a.NotNil(NewBalancer(""))
	a.IsType(&RoundRobin{}, NewBalancer("RoundRobin"))
}
