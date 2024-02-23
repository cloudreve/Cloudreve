package wopi

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
)

type Client interface {
	// NewSession creates a new document session with access token.
	NewSession(uid uint, file *model.File, action ActonType) (*Session, error)
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
	SessionCachePrefix  = "wopi_session_"
	AccessTokenQuery    = "access_token"
	OverwriteHeader     = wopiHeaderPrefix + "Override"
	ServerErrorHeader   = wopiHeaderPrefix + "ServerError"
	RenameRequestHeader = wopiHeaderPrefix + "RequestedName"

	MethodLock        = "LOCK"
	MethodUnlock      = "UNLOCK"
	MethodRefreshLock = "REFRESH_LOCK"
	MethodRename      = "RENAME_FILE"

	wopiSrcPlaceholder    = "WOPI_SOURCE"
	wopiSrcParamDefault   = "WOPISrc"
	languageParamDefault  = "lang"
	sessionExpiresPadding = 10
	wopiHeaderPrefix      = "X-WOPI-"
)

// Init initializes a new global WOPI client.
func Init() {
	settings := model.GetSettingByNames("wopi_endpoint", "wopi_enabled")
	if !model.IsTrueVal(settings["wopi_enabled"]) {
		DefaultMu.Lock()
		Default = nil
		DefaultMu.Unlock()
		return
	}

	cache.Deletes([]string{DiscoverResponseCacheKey}, "")
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

func (c *client) NewSession(uid uint, file *model.File, action ActonType) (*Session, error) {
	if err := c.checkDiscovery(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ext := path.Ext(file.Name)
	availableActions, ok := c.actions[ext]
	if !ok {
		return nil, ErrActionNotSupported
	}

	var (
		actionConfig Action
	)
	fallbackOrder := []ActonType{action, ActionPreview, ActionPreviewFallback, ActionEdit}
	for _, a := range fallbackOrder {
		if actionConfig, ok = availableActions[string(a)]; ok {
			break
		}
	}

	if actionConfig.Urlsrc == "" {
		return nil, ErrActionNotSupported
	}

	// Generate WOPI REST endpoint for given file
	baseURL := model.GetSiteURL()
	linkPath, err := url.Parse(fmt.Sprintf("/api/v3/wopi/files/%s", hashid.HashID(file.ID, hashid.FileID)))
	if err != nil {
		return nil, err
	}

	actionUrl, err := generateActionUrl(actionConfig.Urlsrc, baseURL.ResolveReference(linkPath).String())
	if err != nil {
		return nil, err
	}

	// Create document session
	sessionID := uuid.Must(uuid.NewV4())
	token := util.RandStringRunes(64)
	ttl := model.GetIntSetting("wopi_session_timeout", 36000)
	session := &SessionCache{
		AccessToken: fmt.Sprintf("%s.%s", sessionID, token),
		FileID:      file.ID,
		UserID:      uid,
		Action:      action,
	}
	err = c.cache.Set(SessionCachePrefix+sessionID.String(), *session, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create document session: %w", err)
	}

	sessionRes := &Session{
		AccessToken:    session.AccessToken,
		ActionURL:      actionUrl,
		AccessTokenTTL: time.Now().Add(time.Duration(ttl-sessionExpiresPadding) * time.Second).UnixMilli(),
	}

	return sessionRes, nil
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
	srcReplaced := false
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
			srcReplaced = true
			continue
		}

		queryReplaced.Set(k, queries.Get(k))
	}

	if !srcReplaced {
		queryReplaced.Set(wopiSrcParamDefault, fileSrc)
	}

	// LibreOffice require this flag to show correct language
	queryReplaced.Set(languageParamDefault, "lng")

	actionUrl.RawQuery = queryReplaced.Encode()
	return actionUrl, nil
}
