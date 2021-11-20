package serializer

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSlaveTransferReq_Hash(t *testing.T) {
	a := assert.New(t)
	s1 := &SlaveTransferReq{
		Src:    "1",
		Policy: &model.Policy{},
	}
	s2 := &SlaveTransferReq{
		Src:    "2",
		Policy: &model.Policy{},
	}
	a.NotEqual(s1.Hash("1"), s2.Hash("1"))
}
