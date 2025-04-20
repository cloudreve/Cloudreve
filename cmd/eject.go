package cmd

import (
	"github.com/cloudreve/Cloudreve/v4/application/constants"
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/application/statics"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(ejectCmd)
}

var ejectCmd = &cobra.Command{
	Use:   "eject",
	Short: "Eject all embedded static files",
	Run: func(cmd *cobra.Command, args []string) {
		dep := dependency.NewDependency(
			dependency.WithConfigPath(confPath),
			dependency.WithProFlag(constants.IsPro == "true"),
		)
		logger := dep.Logger()

		if err := statics.Eject(dep.Logger(), dep.Statics()); err != nil {
			logger.Error("Failed to eject static files: %s", err)
			os.Exit(1)
		}
	},
}
