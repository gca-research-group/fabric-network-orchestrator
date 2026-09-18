package cli

import (
	"log/slog"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/spf13/cobra"
)

var chaincodeDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy configured chaincodes",
	Long:  `Deploy configured chaincodes to an already deployed network.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		var config *config.Config
		var err error

		if config, err = LoadConfig(); err != nil {
			return err
		}

		if err := workflows.DeployChaincodes(config); err != nil {
			return err
		}

		slog.Info("Chaincodes deployed successfully.")
		return nil
	},
}

func init() {
	AddConfigCommand(chaincodeDeployCmd)

	chaincodeCmd.AddCommand(chaincodeDeployCmd)
}
