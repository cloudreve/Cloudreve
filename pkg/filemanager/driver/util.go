package driver

import (
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"net/url"
	"path"
	"strings"
)

func ApplyProxyIfNeeded(policy *ent.StoragePolicy, srcUrl *url.URL) (*url.URL, error) {
	// For custom proxy, generate a new proxyed URL:
	// [Proxy Scheme][Proxy Host][Proxy Port][ProxyPath + OriginSrcPath][OriginSrcQuery + ProxyQuery]
	if policy.Settings.CustomProxy {
		proxy, err := url.Parse(policy.Settings.ProxyServer)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		proxy.Path = path.Join(proxy.Path, strings.TrimPrefix(srcUrl.Path, "/"))
		q := proxy.Query()
		if len(q) == 0 {
			proxy.RawQuery = srcUrl.RawQuery
		} else {
			// Merge query parameters
			srcQ := srcUrl.Query()
			for k, _ := range srcQ {
				q.Set(k, srcQ.Get(k))
			}

			proxy.RawQuery = q.Encode()
		}

		srcUrl = proxy
	}

	return srcUrl, nil
}
