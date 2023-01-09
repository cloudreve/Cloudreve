package wopi

import (
	"errors"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"net/url"
	"path"
	"strings"
	"sync"
)

type Client interface {
	// NewSession creates a new document session with access token.
	NewSession(user *model.User, file *model.File, action ActonType) (*Session, error)
	// AvailableExts returns a list of file extensions that are supported by WOPI.
	AvailableExts() []string
}

var (
	ErrActionNotSupported = errors.New("action not supported by current wopi endpoint")

	Default   Client
	DefaultMu sync.Mutex

	queryPlaceholders = map[string]string{
		"BUSINESS_USER":           "",
		"DC_LLCC":                 "lng",
		"DISABLE_ASYNC":           "",
		"DISABLE_CHAT":            "",
		"EMBEDDED":                "true",
		"FULLSCREEN":              "true",
		"HOST_SESSION_ID":         "",
		"SESSION_CONTEXT":         "",
		"RECORDING":               "",
		"THEME_ID":                "darkmode",
		"UI_LLCC":                 "lng",
		"VALIDATOR_TEST_CATEGORY": "",
	}
)

const (
	wopiSrcPlaceholder = "WOPI_SOURCE"
)

// Init initializes a new global WOPI client.
func Init() {
	settings := model.GetSettingByNames("wopi_endpoint", "wopi_enabled")
	if !model.IsTrueVal(settings["wopi_enabled"]) {
		return
	}

	wopiClient, err := NewClient(settings["wopi_endpoint"], cache.Store, request.NewClient())
	if err != nil {
		util.Log().Error("Failed to initialize WOPI client: %s", err)
		return
	}

	DefaultMu.Lock()
	Default = wopiClient
	DefaultMu.Unlock()
}

type client struct {
	cache cache.Driver
	http  request.Client
	mu    sync.RWMutex

	discovery *WopiDiscovery
	actions   map[string]map[string]Action

	config
}

type config struct {
	discoveryEndpoint *url.URL
}

func NewClient(endpoint string, cache cache.Driver, http request.Client) (Client, error) {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WOPI endpoint: %s", err)
	}

	return &client{
		cache: cache,
		http:  http,
		config: config{
			discoveryEndpoint: endpointUrl,
		},
	}, nil
}

func (c *client) NewSession(user *model.User, file *model.File, action ActonType) (*Session, error) {
	if err := c.checkDiscovery(); err != nil {
		return nil, err
	}

	ext := path.Ext(file.Name)
	availableActions, ok := c.actions[ext]
	if !ok {
		return nil, ErrActionNotSupported
	}

	actionConfig, ok := availableActions[string(action)]
	if !ok {
		return nil, ErrActionNotSupported
	}

	actionUrl, err := generateActionUrl(actionConfig.Urlsrc, "")
	if err != nil {
		return nil, err
	}

	fmt.Println(actionUrl)

	return nil, nil
}

// Replace query parameters in action URL template. Some placeholders need to be replaced
// at the frontend, e.g. `THEME_ID`.
func generateActionUrl(src string, fileSrc string) (*url.URL, error) {
	src = strings.ReplaceAll(src, "<", "")
	src = strings.ReplaceAll(src, ">", "")
	actionUrl, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action url: %s", err)
	}

	queries := actionUrl.Query()
	queryReplaced := url.Values{}
	for k := range queries {
		if placeholder, ok := queryPlaceholders[queries.Get(k)]; ok {
			if placeholder != "" {
				queryReplaced.Set(k, placeholder)
			}

			continue
		}

		if queries.Get(k) == wopiSrcPlaceholder {
			queryReplaced.Set(k, fileSrc)
			continue
		}

		queryReplaced.Set(k, queries.Get(k))
	}

	actionUrl.RawQuery = queryReplaced.Encode()
	return actionUrl, nil
}
