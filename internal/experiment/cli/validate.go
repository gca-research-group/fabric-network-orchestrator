package cli

import (
	"fmt"
	"io"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/runner"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var directory string
	var progressInterval int
	command := &cobra.Command{
		Use:   "validate",
		Short: "Validate an existing scenario corpus",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if progressInterval <= 0 {
				return fmt.Errorf("--progress-interval must be greater than zero")
			}
			return runPhase("validation", phaseParameters{Output: directory, ProgressInterval: progressInterval}, cmd.OutOrStdout(), func() (phaseResults, error) {
				return validateDirectory(directory, progressInterval, cmd.OutOrStdout())
			})
		},
	}
	command.Flags().StringVar(&directory, "output", "output", "Directory containing scenarios.parquet; receives results.parquet")
	command.Flags().IntVar(&progressInterval, "progress-interval", defaultProgressInterval, "Report progress every N scenarios (must be greater than zero; start and completion are always reported)")
	return command
}

func validateDirectory(directory string, progressInterval int, stdout io.Writer) (phaseResults, error) {
	total, err := runner.CountScenarios(directory)
	if err != nil {
		return phaseResults{}, err
	}
	reportProgress(stdout, "validation", 0, total, progressInterval)
	summary, err := runner.RunDirectory(directory, func(completed int) {
		reportProgress(stdout, "validation", completed, total, progressInterval)
	})
	if err != nil {
		return phaseResults{}, err
	}

	fmt.Fprintf(stdout, "processed %d: %d passed, %d partial, %d failed\n", summary.Total, summary.Passed, summary.Partial, summary.Failed)
	results := phaseResults{Total: intPointer(summary.Total), Passed: intPointer(summary.Passed), Partial: intPointer(summary.Partial), Failed: intPointer(summary.Failed)}
	if summary.Partial > 0 || summary.Failed > 0 {
		return results, fmt.Errorf("%d scenarios did not produce all expected validation rules", summary.Partial+summary.Failed)
	}

	return results, nil
}
