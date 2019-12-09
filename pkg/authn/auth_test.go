package authn

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	asserts := assert.New(t)

	asserts.NotPanics(func() {
		Init()
	})
	asserts.NotNil(AuthnInstance)
}
