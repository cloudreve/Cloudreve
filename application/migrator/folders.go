package migrator

import (
	"context"
	"fmt"

	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
)

func (m *Migrator) migrateFolders() error {
	m.l.Info("Migrating folders...")
	batchSize := 1000
	// Start from the saved offset if available
	offset := m.state.FolderOffset
	ctx := context.Background()
	foldersCount := 0

	if m.state.FolderIDs == nil {
		m.state.FolderIDs = make(map[int]bool)
	}

	if offset > 0 {
		m.l.Info("Resuming folder migration from offset %d", offset)
	}

	for {
		m.l.Info("Migrating folders with offset %d", offset)
		var folders []model.Folder
		if err := model.DB.Limit(batchSize).Offset(offset).Find(&folders).Error; err != nil {
			return fmt.Errorf("failed to list v3 folders: %w", err)
		}

		if len(folders) == 0 {
			break
		}

		tx, err := m.v4client.Tx(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		batchFoldersCount := 0
		for _, f := range folders {
			if _, ok := m.state.UserIDs[int(f.OwnerID)]; !ok {
				m.l.Warning("Owner ID %d not found, skipping folder %d", f.OwnerID, f.ID)
				continue
			}

			isRoot := f.ParentID == nil
			if isRoot {
				f.Name = ""
			} else if *f.ParentID == 0 {
				m.l.Warning("Parent ID %d not found, skipping folder %d", *f.ParentID, f.ID)
				continue
			}

			stm := tx.File.Create().
				SetRawID(int(f.ID)).
				SetType(int(types.FileTypeFolder)).
				SetCreatedAt(formatTime(f.CreatedAt)).
				SetUpdatedAt(formatTime(f.UpdatedAt)).
				SetName(f.Name).
				SetOwnerID(int(f.OwnerID))

			if _, err := stm.Save(ctx); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to create folder %d: %w", f.ID, err)
			}

			m.state.FolderIDs[int(f.ID)] = true
			m.state.LastFolderID = int(f.ID)

			foldersCount++
			batchFoldersCount++
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Update the offset in state and save after each batch
		offset += batchSize
		m.state.FolderOffset = offset
		if err := m.saveState(); err != nil {
			m.l.Warning("Failed to save state after folder batch: %s", err)
		} else {
			m.l.Info("Saved migration state after processing %d folders in this batch", batchFoldersCount)
		}
	}

	m.l.Info("Successfully migrated %d folders", foldersCount)
	return nil
}

func (m *Migrator) migrateFolderParent() error {
	m.l.Info("Migrating folder parent...")
	batchSize := 1000
	offset := m.state.FolderParentOffset
	ctx := context.Background()

	for {
		m.l.Info("Migrating folder parent with offset %d", offset)
		var folderParents []model.Folder
		if err := model.DB.Limit(batchSize).Offset(offset).Find(&folderParents).Error; err != nil {
			return fmt.Errorf("failed to list v3 folder parents: %w", err)
		}

		if len(folderParents) == 0 {
			break
		}

		tx, err := m.v4client.Tx(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		for _, f := range folderParents {
			if f.ParentID != nil {
				if _, ok := m.state.FolderIDs[int(*f.ParentID)]; !ok {
					m.l.Warning("Folder ID %d not found, skipping folder parent %d", f.ID, f.ID)
					continue
				}

				if _, err := tx.File.UpdateOneID(int(f.ID)).SetParentID(int(*f.ParentID)).Save(ctx); err != nil {
					_ = tx.Rollback()
					return fmt.Errorf("failed to update folder parent %d: %w", f.ID, err)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Update the offset in state and save after each batch
		offset += batchSize
		m.state.FolderParentOffset = offset
		if err := m.saveState(); err != nil {
			m.l.Warning("Failed to save state after folder parent batch: %s", err)
		}
	}

	return nil
}
