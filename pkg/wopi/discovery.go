package wopi

import (
	"encoding/xml"
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
)

type ActonType string

var (
	ActionPreview         = ActonType("embedview")
	ActionPreviewFallback = ActonType("view")
	ActionEdit            = ActonType("edit")
)

func DiscoveryXmlToViewerGroup(xmlStr string) (*setting.ViewerGroup, error) {
	var discovery WopiDiscovery
	if err := xml.Unmarshal([]byte(xmlStr), &discovery); err != nil {
		return nil, fmt.Errorf("failed to parse WOPI discovery XML: %w", err)
	}

	group := &setting.ViewerGroup{
		Viewers: make([]setting.Viewer, 0, len(discovery.NetZone.App)),
	}

	for _, app := range discovery.NetZone.App {
		viewer := setting.Viewer{
			ID:          uuid.Must(uuid.NewV4()).String(),
			DisplayName: app.Name,
			Type:        setting.ViewerTypeWopi,
			Icon:        app.FavIconUrl,
			WopiActions: make(map[string]map[setting.ViewerAction]string),
		}

		for _, action := range app.Action {
			if action.Ext == "" {
				continue
			}

			if _, ok := viewer.WopiActions[action.Ext]; !ok {
				viewer.WopiActions[action.Ext] = make(map[setting.ViewerAction]string)
			}

			if action.Name == string(ActionPreview) {
				viewer.WopiActions[action.Ext][setting.ViewerActionView] = action.Urlsrc
			} else if action.Name == string(ActionPreviewFallback) {
				viewer.WopiActions[action.Ext][setting.ViewerActionView] = action.Urlsrc
			} else if action.Name == string(ActionEdit) {
				viewer.WopiActions[action.Ext][setting.ViewerActionEdit] = action.Urlsrc
			} else if len(viewer.WopiActions[action.Ext]) == 0 {
				delete(viewer.WopiActions, action.Ext)
			}
		}

		viewer.Exts = lo.MapToSlice(viewer.WopiActions, func(key string, value map[setting.ViewerAction]string) string {
			return key
		})

		if len(viewer.WopiActions) > 0 {
			group.Viewers = append(group.Viewers, viewer)
		}
	}

	return group, nil
}
