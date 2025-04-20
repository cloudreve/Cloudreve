package migrator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
	"github.com/cloudreve/Cloudreve/v4/ent/node"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
)

func (m *Migrator) migrateNode() error {
	m.l.Info("Migrating nodes...")

	var nodes []model.Node
	if err := model.DB.Find(&nodes).Error; err != nil {
		return fmt.Errorf("failed to list v3 nodes: %w", err)
	}

	for _, n := range nodes {
		nodeType := node.TypeSlave
		nodeStatus := node.StatusSuspended
		if n.Type == model.MasterNodeType {
			nodeType = node.TypeMaster
		}
		if n.Status == model.NodeActive {
			nodeStatus = node.StatusActive
		}

		cap := &boolset.BooleanSet{}
		settings := &types.NodeSetting{
			Provider: types.DownloaderProviderAria2,
		}

		if n.Aria2Enabled {
			boolset.Sets(map[types.NodeCapability]bool{
				types.NodeCapabilityRemoteDownload: true,
			}, cap)

			aria2Options := &model.Aria2Option{}
			if err := json.Unmarshal([]byte(n.Aria2Options), aria2Options); err != nil {
				return fmt.Errorf("failed to unmarshal aria2 options: %w", err)
			}

			downloaderOptions := map[string]any{}
			if aria2Options.Options != "" {
				if err := json.Unmarshal([]byte(aria2Options.Options), &downloaderOptions); err != nil {
					return fmt.Errorf("failed to unmarshal aria2 options: %w", err)
				}
			}

			settings.Aria2Setting = &types.Aria2Setting{
				Server:   aria2Options.Server,
				Token:    aria2Options.Token,
				Options:  downloaderOptions,
				TempPath: aria2Options.TempPath,
			}
		}

		if n.Type == model.MasterNodeType {
			boolset.Sets(map[types.NodeCapability]bool{
				types.NodeCapabilityExtractArchive: true,
				types.NodeCapabilityCreateArchive:  true,
			}, cap)
		}

		stm := m.v4client.Node.Create().
			SetRawID(int(n.ID)).
			SetCreatedAt(formatTime(n.CreatedAt)).
			SetUpdatedAt(formatTime(n.UpdatedAt)).
			SetName(n.Name).
			SetType(nodeType).
			SetStatus(nodeStatus).
			SetServer(n.Server).
			SetSlaveKey(n.SlaveKey).
			SetCapabilities(cap).
			SetSettings(settings).
			SetWeight(n.Rank)

		if err := stm.Exec(context.Background()); err != nil {
			return fmt.Errorf("failed to create node %q: %w", n.Name, err)
		}

	}

	return nil
}
