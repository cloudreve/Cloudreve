package wopi

import (
	"encoding/gob"
	"encoding/xml"
	"net/url"
)

// Response content from discovery endpoint.
type WopiDiscovery struct {
	XMLName xml.Name `xml:"wopi-discovery"`
	Text    string   `xml:",chardata"`
	NetZone struct {
		Text string `xml:",chardata"`
		Name string `xml:"name,attr"`
		App  []struct {
			Text                 string   `xml:",chardata"`
			Name                 string   `xml:"name,attr"`
			FavIconUrl           string   `xml:"favIconUrl,attr"`
			BootstrapperUrl      string   `xml:"bootstrapperUrl,attr"`
			AppBootstrapperUrl   string   `xml:"appBootstrapperUrl,attr"`
			ApplicationBaseUrl   string   `xml:"applicationBaseUrl,attr"`
			StaticResourceOrigin string   `xml:"staticResourceOrigin,attr"`
			CheckLicense         string   `xml:"checkLicense,attr"`
			Action               []Action `xml:"action"`
		} `xml:"app"`
	} `xml:"net-zone"`
	ProofKey struct {
		Text        string `xml:",chardata"`
		Oldvalue    string `xml:"oldvalue,attr"`
		Oldmodulus  string `xml:"oldmodulus,attr"`
		Oldexponent string `xml:"oldexponent,attr"`
		Value       string `xml:"value,attr"`
		Modulus     string `xml:"modulus,attr"`
		Exponent    string `xml:"exponent,attr"`
	} `xml:"proof-key"`
}

type Action struct {
	Text      string `xml:",chardata"`
	Name      string `xml:"name,attr"`
	Ext       string `xml:"ext,attr"`
	Default   string `xml:"default,attr"`
	Urlsrc    string `xml:"urlsrc,attr"`
	Requires  string `xml:"requires,attr"`
	Targetext string `xml:"targetext,attr"`
	Progid    string `xml:"progid,attr"`
	UseParent string `xml:"useParent,attr"`
	Newprogid string `xml:"newprogid,attr"`
	Newext    string `xml:"newext,attr"`
}

type Session struct {
	AccessToken    string
	AccessTokenTTL int64
	ActionURL      *url.URL
}

type SessionCache struct {
	AccessToken string
	FileID      uint
	UserID      uint
	Action      ActonType
}

func init() {
	gob.Register(WopiDiscovery{})
	gob.Register(Action{})
	gob.Register(SessionCache{})
}
