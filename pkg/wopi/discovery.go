package wopi

import (
	"encoding/xml"
	"fmt"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"net/http"
	"strings"
)

type ActonType string

var (
	ActionPreview         = ActonType("embedview")
	ActionPreviewFallback = ActonType("view")
	ActionEdit            = ActonType("edit")
)

const (
	DiscoverResponseCacheKey = "wopi_discover"
	DiscoverRefreshDuration  = 24 * 3600 // 24 hrs
)

func (c *client) AvailableExts() []string {
	if err := c.checkDiscovery(); err != nil {
		util.Log().Error("Failed to check WOPI discovery: %s", err)
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	exts := make([]string, 0, len(c.actions))
	for ext, actions := range c.actions {
		_, previewable := actions[string(ActionPreview)]
		_, editable := actions[string(ActionEdit)]
		_, previewableFallback := actions[string(ActionPreviewFallback)]

		if previewable || editable || previewableFallback {
			exts = append(exts, strings.TrimPrefix(ext, "."))
		}
	}

	return exts
}

// checkDiscovery checks if discovery content is needed to be refreshed.
// If so, it will refresh discovery content.
func (c *client) checkDiscovery() error {
	c.mu.RLock()
	if c.discovery == nil {
		c.mu.RUnlock()
		return c.refreshDiscovery()
	}

	c.mu.RUnlock()
	return nil
}

// refresh Discovery action configs.
func (c *client) refreshDiscovery() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cached, exist := c.cache.Get(DiscoverResponseCacheKey)
	if exist {
		cachedDiscovery := cached.(WopiDiscovery)
		c.discovery = &cachedDiscovery
	} else {
		res, err := c.http.Request("GET", c.config.discoveryEndpoint.String(), nil).
			CheckHTTPResponse(http.StatusOK).GetResponse()
		if err != nil {
			return fmt.Errorf("failed to request discovery endpoint: %w", err)
		}

		if err := xml.Unmarshal([]byte(res), &c.discovery); err != nil {
			return fmt.Errorf("failed to parse response discovery endpoint: %w", err)
		}

		if err := c.cache.Set(DiscoverResponseCacheKey, *c.discovery, DiscoverRefreshDuration); err != nil {
			return err
		}
	}

	// construct actions map
	c.actions = make(map[string]map[string]Action)
	for _, app := range c.discovery.NetZone.App {
		for _, action := range app.Action {
			if action.Ext == "" {
				continue
			}

			if _, ok := c.actions["."+action.Ext]; !ok {
				c.actions["."+action.Ext] = make(map[string]Action)
			}

			c.actions["."+action.Ext][action.Name] = action
		}
	}

	return nil
}
