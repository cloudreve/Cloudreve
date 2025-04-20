package cmd

import (
	"os"
	"path/filepath"

	"github.com/cloudreve/Cloudreve/v4/application/constants"
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/application/migrator"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/spf13/cobra"
)

var (
	v3ConfPath string
	forceReset bool
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.PersistentFlags().StringVar(&v3ConfPath, "v3-conf", "", "Path to the v3 config file")
	migrateCmd.PersistentFlags().BoolVar(&forceReset, "force-reset", false, "Force reset migration state and start from beginning")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from v3 to v4",
	Run: func(cmd *cobra.Command, args []string) {
		dep := dependency.NewDependency(
			dependency.WithConfigPath(confPath),
			dependency.WithRequiredDbVersion(constants.BackendVersion),
			dependency.WithProFlag(constants.IsPro == "true"),
		)
		logger := dep.Logger()
		logger.Info("Migrating from v3 to v4...")

		if v3ConfPath == "" {
			logger.Error("v3 config file is required, please use -v3-conf to specify the path.")
			os.Exit(1)
		}

		// Check if state file exists and warn about resuming
		stateFilePath := filepath.Join(filepath.Dir(v3ConfPath), "migration_state.json")
		if util.Exists(stateFilePath) && !forceReset {
			logger.Info("Found existing migration state file at %s. Migration will resume from the last successful step.", stateFilePath)
			logger.Info("If you want to start migration from the beginning, please use --force-reset flag.")
		} else if forceReset && util.Exists(stateFilePath) {
			logger.Info("Force resetting migration state. Will start from the beginning.")
			if err := os.Remove(stateFilePath); err != nil {
				logger.Error("Failed to remove migration state file: %s", err)
				os.Exit(1)
			}
		}

		migrator, err := migrator.NewMigrator(dep, v3ConfPath)
		if err != nil {
			logger.Error("Failed to create migrator: %s", err)
			os.Exit(1)
		}

		if err := migrator.Migrate(); err != nil {
			logger.Error("Failed to migrate: %s", err)
			logger.Info("Migration failed but state has been saved. You can retry with the same command to resume from the last successful step.")
			os.Exit(1)
		}

		logger.Info("Migration from v3 to v4 completed successfully.")
	},
}
