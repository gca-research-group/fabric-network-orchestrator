package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
	"github.com/spf13/cobra"
)

func newGenerateCommand() *cobra.Command {
	var seedPath, directory string
	var mutationCount int
	var progressInterval int
	command := &cobra.Command{
		Use:   "generate",
		Short: "Generate scenarios from a seed YAML configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if progressInterval <= 0 {
				return errors.New("--progress-interval must be greater than zero")
			}
			if seedPath == "" {
				return errors.New("missing required flag: --seed")
			}
			stdout := cmd.OutOrStdout()
			return runPhase("generation", phaseParameters{Seed: seedPath, MutationCount: &mutationCount, Output: directory, ProgressInterval: progressInterval}, stdout, func() (phaseResults, error) {
				seedYAML, err := os.ReadFile(seedPath)
				if err != nil {
					return phaseResults{}, fmt.Errorf("read seed YAML %q: %w", seedPath, err)
				}
				summary, err := generator.Generate(seedYAML, mutationCount, directory, func(completed, total int) {
					reportProgress(stdout, "generation", completed, total, progressInterval)
				})
				if err != nil {
					return phaseResults{}, err
				}
				fmt.Fprintf(stdout, "generated %d scenarios\n", summary.Total)
				return phaseResults{Total: intPointer(summary.Total), MutationOperatorsUsed: intPointer(summary.MutationOperatorsUsed)}, nil
			})
		},
	}
	command.Flags().StringVar(&seedPath, "seed", "", "Path to the seed YAML configuration (required)")
	command.Flags().IntVar(&mutationCount, "mutation-count", generator.DefaultMutationCount, "Number of mutations per scenario")
	command.Flags().IntVar(&progressInterval, "progress-interval", defaultProgressInterval, "Report progress every N scenarios (must be greater than zero; start and completion are always reported)")
	command.Flags().StringVar(&directory, "output", "output", "Directory for generated scenarios")
	command.MarkFlagFilename("seed", "yaml", "yml")
	command.MarkFlagRequired("seed")
	return command
}

func intPointer(value int) *int { return &value }
