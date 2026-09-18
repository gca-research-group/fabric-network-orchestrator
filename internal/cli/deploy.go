package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Generate artifacts and deploy the network and chaincodes",
	Long:  "Generate artifacts, deploy the network, and deploy configured chaincodes in sequence. Requires an empty artifact output directory and stops at the first failure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		if err := workflows.Deploy(cfg); err != nil {
			return err
		}
		slog.Info("Deployment completed successfully.")
		return nil
	},
}

func init() {
	AddConfigCommand(deployCmd)
	rootCmd.AddCommand(deployCmd)
}
