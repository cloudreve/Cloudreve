package migrator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudreve/Cloudreve/v4/pkg/util"
)

func migrateAvatars(m *Migrator) error {
	m.l.Info("Migrating avatars files...")
	avatarRoot := util.RelativePath(m.state.V3AvatarPath)

	for uid, _ := range m.state.UserIDs {
		avatarPath := filepath.Join(avatarRoot, fmt.Sprintf("avatar_%d_2.png", uid))

		// check if file exists
		if util.Exists(avatarPath) {
			m.l.Info("Migrating avatar for user %d", uid)
			// Copy to v4 avatar path
			v4Path := filepath.Join(util.DataPath("avatar"), fmt.Sprintf("avatar_%d.png", uid))

			// copy
			origin, err := os.Open(avatarPath)
			if err != nil {
				return fmt.Errorf("failed to open avatar file: %w", err)
			}
			defer origin.Close()

			dest, err := util.CreatNestedFile(v4Path)
			if err != nil {
				return fmt.Errorf("failed to create avatar file: %w", err)
			}
			defer dest.Close()

			_, err = io.Copy(dest, origin)

			if err != nil {
				m.l.Warning("Failed to copy avatar file: %s, skipping...", err)
			}
		}
	}

	return nil
}
