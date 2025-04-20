package migrator

import (
	"context"
	"fmt"

	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
	"github.com/cloudreve/Cloudreve/v4/ent/file"
	"github.com/cloudreve/Cloudreve/v4/pkg/conf"
)

func (m *Migrator) migrateDirectLink() error {
	m.l.Info("Migrating direct links...")
	batchSize := 1000
	offset := m.state.DirectLinkOffset
	ctx := context.Background()

	if m.state.DirectLinkOffset > 0 {
		m.l.Info("Resuming direct link migration from offset %d", offset)
	}

	for {
		m.l.Info("Migrating direct links with offset %d", offset)
		var directLinks []model.SourceLink
		if err := model.DB.Limit(batchSize).Offset(offset).Find(&directLinks).Error; err != nil {
			return fmt.Errorf("failed to list v3 direct links: %w", err)
		}

		if len(directLinks) == 0 {
			if m.dep.ConfigProvider().Database().Type == conf.PostgresDB {
				m.l.Info("Resetting direct link ID sequence for postgres...")
				m.v4client.DirectLink.ExecContext(ctx, "SELECT SETVAL('direct_links_id_seq',  (SELECT MAX(id) FROM direct_links))")
			}
			break
		}

		tx, err := m.v4client.Tx(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		for _, dl := range directLinks {
			sourceId := int(dl.FileID) + m.state.LastFolderID
			// check if file exists
			_, err = tx.File.Query().Where(file.ID(sourceId)).First(ctx)
			if err != nil {
				m.l.Warning("File %d not found, skipping direct link %d", sourceId, dl.ID)
				continue
			}

			stm := tx.DirectLink.Create().
				SetCreatedAt(formatTime(dl.CreatedAt)).
				SetUpdatedAt(formatTime(dl.UpdatedAt)).
				SetRawID(int(dl.ID)).
				SetFileID(sourceId).
				SetName(dl.Name).
				SetDownloads(dl.Downloads).
				SetSpeed(0)

			if _, err := stm.Save(ctx); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to create direct link %d: %w", dl.ID, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		offset += batchSize
		m.state.DirectLinkOffset = offset
		if err := m.saveState(); err != nil {
			m.l.Warning("Failed to save state after direct link batch: %s", err)
		} else {
			m.l.Info("Saved migration state after processing this batch")
		}
	}

	return nil

}
