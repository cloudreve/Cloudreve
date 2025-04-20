package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/conf"
)

func (m *Migrator) migrateFile() error {
	m.l.Info("Migrating files...")
	batchSize := 1000
	offset := m.state.FileOffset
	ctx := context.Background()

	if m.state.FileConflictRename == nil {
		m.state.FileConflictRename = make(map[uint]string)
	}

	if m.state.EntitySources == nil {
		m.state.EntitySources = make(map[string]int)
	}

	if offset > 0 {
		m.l.Info("Resuming file migration from offset %d", offset)
	}

out:
	for {
		m.l.Info("Migrating files with offset %d", offset)
		var files []model.File
		if err := model.DB.Limit(batchSize).Offset(offset).Find(&files).Error; err != nil {
			return fmt.Errorf("failed to list v3 files: %w", err)
		}

		if len(files) == 0 {
			if m.dep.ConfigProvider().Database().Type == conf.PostgresDB {
				m.l.Info("Resetting file ID sequence for postgres...")
				m.v4client.File.ExecContext(ctx, "SELECT SETVAL('files_id_seq',  (SELECT MAX(id) FROM files))")
			}
			break
		}

		tx, err := m.v4client.Tx(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		for _, f := range files {
			if _, ok := m.state.FolderIDs[int(f.FolderID)]; !ok {
				m.l.Warning("Folder ID %d for file %d not found, skipping", f.FolderID, f.ID)
				continue
			}

			if _, ok := m.state.UserIDs[int(f.UserID)]; !ok {
				m.l.Warning("User ID %d for file %d not found, skipping", f.UserID, f.ID)
				continue
			}

			if _, ok := m.state.PolicyIDs[int(f.PolicyID)]; !ok {
				m.l.Warning("Policy ID %d for file %d not found, skipping", f.PolicyID, f.ID)
				continue
			}

			metadata := make(map[string]string)
			if f.Metadata != "" {
				json.Unmarshal([]byte(f.Metadata), &metadata)
			}

			var (
				thumbnail *ent.Entity
				entity    *ent.Entity
				err       error
			)

			if metadata[model.ThumbStatusMetadataKey] == model.ThumbStatusExist {
				size := int64(0)
				if m.state.LocalPolicyIDs[int(f.PolicyID)] {
					thumbFile, err := os.Stat(f.SourceName + m.state.ThumbSuffix)
					if err == nil {
						size = thumbFile.Size()
					}
					m.l.Warning("Thumbnail file %s for file %d not found, use 0 size", f.SourceName+m.state.ThumbSuffix, f.ID)
				}
				// Insert thumbnail entity
				thumbnail, err = m.insertEntity(tx, f.SourceName+m.state.ThumbSuffix, int(types.EntityTypeThumbnail), int(f.PolicyID), int(f.UserID), size)
				if err != nil {
					_ = tx.Rollback()
					return fmt.Errorf("failed to insert thumbnail entity: %w", err)
				}
			}

			// Insert file version entity
			entity, err = m.insertEntity(tx, f.SourceName, int(types.EntityTypeVersion), int(f.PolicyID), int(f.UserID), int64(f.Size))
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to insert file version entity: %w", err)
			}

			fname := f.Name
			if _, ok := m.state.FileConflictRename[f.ID]; ok {
				fname = m.state.FileConflictRename[f.ID]
			}

			stm := tx.File.Create().
				SetCreatedAt(formatTime(f.CreatedAt)).
				SetUpdatedAt(formatTime(f.UpdatedAt)).
				SetName(fname).
				SetRawID(int(f.ID) + m.state.LastFolderID).
				SetOwnerID(int(f.UserID)).
				SetSize(int64(f.Size)).
				SetPrimaryEntity(entity.ID).
				SetFileChildren(int(f.FolderID)).
				SetType(int(types.FileTypeFile)).
				SetStoragePoliciesID(int(f.PolicyID)).
				AddEntities(entity)

			if thumbnail != nil {
				stm.AddEntities(thumbnail)
			}

			if _, err := stm.Save(ctx); err != nil {
				_ = tx.Rollback()
				if ent.IsConstraintError(err) {
					if _, ok := m.state.FileConflictRename[f.ID]; ok {
						return fmt.Errorf("file %d already exists, but new name is already in conflict rename map, please resolve this manually", f.ID)
					}

					m.l.Warning("File %d already exists, will retry with new name in next batch", f.ID)
					m.state.FileConflictRename[f.ID] = fmt.Sprintf("%d_%s", f.ID, f.Name)
					continue out
				}
				return fmt.Errorf("failed to create file %d: %w", f.ID, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		offset += batchSize
		m.state.FileOffset = offset
		if err := m.saveState(); err != nil {
			m.l.Warning("Failed to save state after file batch: %s", err)
		} else {
			m.l.Info("Saved migration state after processing this batch")
		}
	}

	return nil
}

func (m *Migrator) insertEntity(tx *ent.Tx, source string, entityType, policyID, createdBy int, size int64) (*ent.Entity, error) {

	// find existing one
	entityKey := strconv.Itoa(policyID) + "+" + source
	if existingId, ok := m.state.EntitySources[entityKey]; ok {
		existing, err := tx.Entity.UpdateOneID(existingId).
			AddReferenceCount(1).
			Save(context.Background())
		if err == nil {
			return existing, nil
		}
		m.l.Warning("Failed to update existing entity %d: %s, fallback to create new one.", existingId, err)
	}

	// create new one
	e, err := tx.Entity.Create().
		SetSource(source).
		SetType(entityType).
		SetSize(size).
		SetStoragePolicyEntities(policyID).
		SetCreatedBy(createdBy).
		SetReferenceCount(1).
		Save(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create new entity: %w", err)
	}

	m.state.EntitySources[entityKey] = e.ID
	return e, nil
}
