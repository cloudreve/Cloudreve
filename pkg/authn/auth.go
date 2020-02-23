package authn

import (
	model "github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/util"
	"github.com/duo-labs/webauthn/webauthn"
	"sync"
)

var AuthnInstance *webauthn.WebAuthn
var Lock sync.RWMutex

// Init 初始化webauthn
func Init() {
	Lock.Lock()
	defer Lock.Unlock()
	var err error
	base := model.GetSiteURL()
	AuthnInstance, err = webauthn.New(&webauthn.Config{
		RPDisplayName: model.GetSettingByName("siteName"), // Display Name for your site
		RPID:          base.Hostname(),                    // Generally the FQDN for your site
		RPOrigin:      base.String(),                      // The origin URL for WebAuthn requests
	})
	if err != nil {
		util.Log().Error("无法初始化WebAuthn, %s", err)
	}
}
