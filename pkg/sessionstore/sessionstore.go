package sessionstore

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/gin-contrib/sessions"
)

type Store interface {
	sessions.Store
}

func NewStore(driver cache.Driver, keyPairs ...[]byte) Store {
	return &store{newKvStore("cd_session_", driver, keyPairs...)}
}

type store struct {
	*kvStore
}

func (c *store) Options(options sessions.Options) {
	c.kvStore.Options = options.ToGorillaOptions()
}
