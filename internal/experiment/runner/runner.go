package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
)

type Status string

const (
	StatusPassed  Status = "passed"
	StatusPartial Status = "partial"
	StatusFailed  Status = "failed"
)

type Result struct {
	Scenario string
	Expected []validate.RuleID
	Actual   []validate.RuleID
	Missing  []validate.RuleID
	Status   Status
	Error    string
}

type resultRow struct {
	Scenario string            `parquet:"scenario,dict"`
	Expected []validate.RuleID `parquet:"expectedRules,list"`
	Actual   []validate.RuleID `parquet:"actualRules,list"`
	Missing  []validate.RuleID `parquet:"missingRules,list"`
	Status   Status            `parquet:"status,dict"`
	Error    *string           `parquet:"error,optional"`
}

func parquetResult(result Result) resultRow {
	row := resultRow{Scenario: result.Scenario, Expected: result.Expected, Actual: result.Actual, Missing: result.Missing, Status: result.Status}
	if result.Error != "" {
		row.Error = &result.Error
	}
	return row
}

type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Partial int `json:"partial"`
	Failed  int `json:"failed"`
}

type ProgressFunc func(completed int)

func RunDirectory(outputDirectory string, progress ProgressFunc) (Summary, error) {
	manifest, err := openScenarioManifest(outputDirectory)
	if err != nil {
		return Summary{}, err
	}
	defer manifest.Close()

	reader := parquet.NewGenericReader[generator.ScenarioRules](manifest)
	defer reader.Close()

	resultsFile, err := os.CreateTemp(outputDirectory, ".results-*.parquet")
	if err != nil {
		return Summary{}, fmt.Errorf("create result document: %w", err)
	}
	defer os.Remove(resultsFile.Name())
	writer := parquet.NewGenericWriter[resultRow](resultsFile,
		parquet.Compression(&zstd.Codec{}),
		parquet.MaxRowsPerRowGroup(64*1024),
	)

	summary := Summary{}
	rows := make([]generator.ScenarioRules, 1024)
	for {
		n, readErr := reader.Read(rows)
		for _, scenario := range rows[:n] {
			if err := validateScenario(scenario); err != nil {
				return summary, errors.Join(fmt.Errorf("decode scenario manifest entry: %w", err), writer.Close(), resultsFile.Close())
			}

			result := evaluate(outputDirectory, scenario)
			if _, err := writer.Write([]resultRow{parquetResult(result)}); err != nil {
				return summary, errors.Join(fmt.Errorf("write result document: %w", err), writer.Close(), resultsFile.Close())
			}

			summary.Total++
			switch result.Status {
			case StatusPassed:
				summary.Passed++
			case StatusPartial:
				summary.Partial++
			case StatusFailed:
				summary.Failed++
			}
			if progress != nil {
				progress(summary.Total)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return summary, errors.Join(fmt.Errorf("decode scenario manifest entry: %w", readErr), writer.Close(), resultsFile.Close())
		}
	}
	if err := errors.Join(writer.Close(), resultsFile.Close()); err != nil {
		return summary, fmt.Errorf("close result document: %w", err)
	}
	if err := os.Rename(resultsFile.Name(), filepath.Join(outputDirectory, "results.parquet")); err != nil {
		return summary, fmt.Errorf("replace result document: %w", err)
	}

	return summary, nil
}

func evaluate(outputDirectory string, scenario generator.ScenarioRules) Result {
	expected := expectedRules(scenario)
	result := Result{Scenario: scenario.Scenario, Expected: expected}
	path := filepath.Join(outputDirectory, "config", scenario.Scenario+".yaml")
	_, err := config.LoadConfigFromPath(path)

	validationErrors := validate.Errors(err)
	if len(validationErrors) > 0 {
		for _, validationError := range validationErrors {
			result.Actual = appendRuleOnce(result.Actual, validationError.RuleID)
		}
		result.Missing = missingRules(result.Actual, expected)
		result.Status = classify(len(expected), len(result.Missing))
	} else if err == nil {
		result.Status = StatusFailed
		result.Missing = append(result.Missing, expected...)
		result.Error = "configuration unexpectedly passed validation"
	} else {
		result.Status = StatusFailed
		result.Missing = append(result.Missing, expected...)
		result.Error = err.Error()
	}
	return result
}

func expectedRules(scenario generator.ScenarioRules) []validate.RuleID {
	rules := make([]validate.RuleID, 0, len(scenario.Mutations))
	for _, mutation := range scenario.Mutations {
		rules = append(rules, mutation.Rule)
	}
	return rules
}

func validateScenario(scenario generator.ScenarioRules) error {
	if scenario.Scenario == "" {
		return fmt.Errorf("scenario is required")
	}
	if len(scenario.Mutations) == 0 {
		return fmt.Errorf("scenario %s: mutations are required", scenario.Scenario)
	}
	for _, mutation := range scenario.Mutations {
		if mutation.Rule == "" {
			return fmt.Errorf("scenario %s: mutation rule is required", scenario.Scenario)
		}
		if mutation.OperatorIndex < 0 {
			return fmt.Errorf("scenario %s: mutation operator index must not be negative", scenario.Scenario)
		}
		if _, found := generator.FindMutationOperator(mutation); !found {
			return fmt.Errorf("scenario %s: unknown mutation operator %s/%d", scenario.Scenario, mutation.Rule, mutation.OperatorIndex)
		}
	}
	return nil
}

func missingRules(actual, expected []validate.RuleID) []validate.RuleID {
	var missing []validate.RuleID
	for _, expectedRule := range expected {
		found := false
		for _, actualRule := range actual {
			if actualRule == expectedRule {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, expectedRule)
		}
	}
	return missing
}

func classify(expected, missing int) Status {
	if missing == 0 {
		return StatusPassed
	}
	if missing < expected {
		return StatusPartial
	}
	return StatusFailed
}

func appendRuleOnce(rules []validate.RuleID, ruleID validate.RuleID) []validate.RuleID {
	for _, existing := range rules {
		if existing == ruleID {
			return rules
		}
	}
	return append(rules, ruleID)
}
