package migrator

import (
	"context"
	"fmt"

	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
	"github.com/cloudreve/Cloudreve/v4/pkg/conf"
)

func (m *Migrator) migrateWebdav() error {
	m.l.Info("Migrating webdav accounts...")

	batchSize := 1000
	offset := m.state.WebdavOffset
	ctx := context.Background()

	if m.state.WebdavOffset > 0 {
		m.l.Info("Resuming webdav migration from offset %d", offset)
	}

	for {
		m.l.Info("Migrating webdav accounts with offset %d", offset)
		var webdavAccounts []model.Webdav
		if err := model.DB.Limit(batchSize).Offset(offset).Find(&webdavAccounts).Error; err != nil {
			return fmt.Errorf("failed to list v3 webdav accounts: %w", err)
		}

		if len(webdavAccounts) == 0 {
			if m.dep.ConfigProvider().Database().Type == conf.PostgresDB {
				m.l.Info("Resetting webdav account ID sequence for postgres...")
				m.v4client.DavAccount.ExecContext(ctx, "SELECT SETVAL('dav_accounts_id_seq',  (SELECT MAX(id) FROM dav_accounts))")
			}
			break
		}

		tx, err := m.v4client.Tx(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		for _, webdavAccount := range webdavAccounts {
			if _, ok := m.state.UserIDs[int(webdavAccount.UserID)]; !ok {
				m.l.Warning("User %d not found, skipping webdav account %d", webdavAccount.UserID, webdavAccount.ID)
				continue
			}

			props := types.DavAccountProps{}
			options := boolset.BooleanSet{}

			if webdavAccount.Readonly {
				boolset.Set(int(types.DavAccountReadOnly), true, &options)
			}

			if webdavAccount.UseProxy {
				boolset.Set(int(types.DavAccountProxy), true, &options)
			}

			stm := tx.DavAccount.Create().
				SetCreatedAt(formatTime(webdavAccount.CreatedAt)).
				SetUpdatedAt(formatTime(webdavAccount.UpdatedAt)).
				SetRawID(int(webdavAccount.ID)).
				SetName(webdavAccount.Name).
				SetURI("cloudreve://my" + webdavAccount.Root).
				SetPassword(webdavAccount.Password).
				SetProps(&props).
				SetOptions(&options).
				SetOwnerID(int(webdavAccount.UserID))

			if _, err := stm.Save(ctx); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to create webdav account %d: %w", webdavAccount.ID, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		offset += batchSize
		m.state.WebdavOffset = offset
		if err := m.saveState(); err != nil {
			m.l.Warning("Failed to save state after webdav batch: %s", err)
		} else {
			m.l.Info("Saved migration state after processing this batch")
		}
	}

	return nil
}
