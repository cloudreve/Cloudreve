package balancer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoundRobin_NextIndex(t *testing.T) {
	a := assert.New(t)
	r := &RoundRobin{}
	total := 5
	for i := 1; i < total; i++ {
		a.Equal(i, r.NextIndex(total))
	}
	for i := 0; i < total; i++ {
		a.Equal(i, r.NextIndex(total))
	}
}

func TestRoundRobin_NextPeer(t *testing.T) {
	a := assert.New(t)
	r := &RoundRobin{}

	// not slice
	{
		err, _ := r.NextPeer("s")
		a.Equal(ErrInputNotSlice, err)
	}

	// no nodes
	{
		err, _ := r.NextPeer([]string{})
		a.Equal(ErrNoAvaliableNode, err)
	}

	// pass
	{
		err, res := r.NextPeer([]string{"a"})
		a.NoError(err)
		a.Equal("a", res.(string))
	}
}
