package authn

import (
	"fmt"
	"github.com/duo-labs/webauthn/webauthn"
)

var Authn *webauthn.WebAuthn

func Init() {
	var err error
	Authn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "Duo Labs",                 // Display Name for your site
		RPID:          "localhost",                // Generally the FQDN for your site
		RPOrigin:      "http://localhost:3000",    // The origin URL for WebAuthn requests
		RPIcon:        "https://duo.com/logo.png", // Optional icon URL for your site
	})
	if err != nil {
		fmt.Println(err)
	}
}
