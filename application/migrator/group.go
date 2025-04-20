package migrator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
	"github.com/cloudreve/Cloudreve/v4/pkg/conf"
	"github.com/samber/lo"
)

func (m *Migrator) migrateGroup() error {
	m.l.Info("Migrating groups...")

	var groups []model.Group
	if err := model.DB.Find(&groups).Error; err != nil {
		return fmt.Errorf("failed to list v3 groups: %w", err)
	}

	for _, group := range groups {
		cap := &boolset.BooleanSet{}
		var (
			opts     model.GroupOption
			policies []int
		)
		if err := json.Unmarshal([]byte(group.Options), &opts); err != nil {
			return fmt.Errorf("failed to unmarshal options for group %q: %w", group.Name, err)
		}

		if err := json.Unmarshal([]byte(group.Policies), &policies); err != nil {
			return fmt.Errorf("failed to unmarshal policies for group %q: %w", group.Name, err)
		}

		policies = lo.Filter(policies, func(id int, _ int) bool {
			_, exist := m.state.PolicyIDs[id]
			return exist
		})

		newOpts := &types.GroupSetting{
			CompressSize:          int64(opts.CompressSize),
			DecompressSize:        int64(opts.DecompressSize),
			RemoteDownloadOptions: opts.Aria2Options,
			SourceBatchSize:       opts.SourceBatchSize,
			RedirectedSource:      opts.RedirectedSource,
			Aria2BatchSize:        opts.Aria2BatchSize,
			MaxWalkedFiles:        100000,
			TrashRetention:        7 * 24 * 3600,
		}

		boolset.Sets(map[types.GroupPermission]bool{
			types.GroupPermissionIsAdmin:          group.ID == 1,
			types.GroupPermissionIsAnonymous:      group.ID == 3,
			types.GroupPermissionShareDownload:    opts.ShareDownload,
			types.GroupPermissionWebDAV:           group.WebDAVEnabled,
			types.GroupPermissionArchiveDownload:  opts.ArchiveDownload,
			types.GroupPermissionArchiveTask:      opts.ArchiveTask,
			types.GroupPermissionWebDAVProxy:      opts.WebDAVProxy,
			types.GroupPermissionRemoteDownload:   opts.Aria2,
			types.GroupPermissionAdvanceDelete:    opts.AdvanceDelete,
			types.GroupPermissionShare:            group.ShareEnabled,
			types.GroupPermissionRedirectedSource: opts.RedirectedSource,
		}, cap)

		stm := m.v4client.Group.Create().
			SetRawID(int(group.ID)).
			SetCreatedAt(formatTime(group.CreatedAt)).
			SetUpdatedAt(formatTime(group.UpdatedAt)).
			SetName(group.Name).
			SetMaxStorage(int64(group.MaxStorage)).
			SetSpeedLimit(group.SpeedLimit).
			SetPermissions(cap).
			SetSettings(newOpts)

		if len(policies) > 0 {
			stm.SetStoragePoliciesID(policies[0])
		}

		if _, err := stm.Save(context.Background()); err != nil {
			return fmt.Errorf("failed to create group %q: %w", group.Name, err)
		}
	}

	if m.dep.ConfigProvider().Database().Type == conf.PostgresDB {
		m.l.Info("Resetting group ID sequence for postgres...")
		m.v4client.Group.ExecContext(context.Background(), "SELECT SETVAL('groups_id_seq',  (SELECT MAX(id) FROM groups))")
	}

	return nil
}
