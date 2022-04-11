package serializer

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildObjectList(t *testing.T) {
	a := assert.New(t)
	res := BuildObjectList(1, []Object{{}, {}}, &model.Policy{})
	a.NotEmpty(res.Parent)
	a.NotNil(res.Policy)
	a.Len(res.Objects, 2)
}
