package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	Scenario string            `json:"scenario"`
	Expected []validate.RuleID `json:"expectedRules"`
	Actual   []validate.RuleID `json:"actualRules,omitempty"`
	Missing  []validate.RuleID `json:"missingRules,omitempty"`
	Status   Status            `json:"status"`
	Error    string            `json:"error,omitempty"`
}

type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Partial int `json:"partial"`
	Failed  int `json:"failed"`
}

type ProgressFunc func(completed int)

func RunDirectory(outputDirectory string, progress ProgressFunc) (Summary, error) {
	manifest, err := os.Open(filepath.Join(outputDirectory, "scenarios.json"))
	if err != nil {
		return Summary{}, fmt.Errorf("open scenario manifest: %w", err)
	}
	defer manifest.Close()

	decoder := json.NewDecoder(bufio.NewReader(manifest))
	token, err := decoder.Token()
	if err != nil {
		return Summary{}, fmt.Errorf("decode scenario manifest: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return Summary{}, fmt.Errorf("decode scenario manifest: expected JSON array")
	}

	resultsFile, err := os.Create(filepath.Join(outputDirectory, "results.json"))
	if err != nil {
		return Summary{}, fmt.Errorf("create result document: %w", err)
	}
	writer := bufio.NewWriter(resultsFile)
	if _, err := writer.WriteString("{\n  \"results\": [\n"); err != nil {
		resultsFile.Close()
		return Summary{}, fmt.Errorf("write result document: %w", err)
	}

	summary := Summary{}
	for decoder.More() {
		var scenario generator.ScenarioRules
		if err := decoder.Decode(&scenario); err != nil {
			resultsFile.Close()
			return summary, fmt.Errorf("decode scenario manifest entry: %w", err)
		}
		if err := validateScenario(scenario); err != nil {
			resultsFile.Close()
			return summary, fmt.Errorf("decode scenario manifest entry: %w", err)
		}

		result := evaluate(outputDirectory, scenario)
		data, err := json.Marshal(result)
		if err != nil {
			resultsFile.Close()
			return summary, fmt.Errorf("encode result for scenario %s: %w", scenario.Scenario, err)
		}
		if summary.Total > 0 {
			if _, err := writer.WriteString(",\n"); err != nil {
				resultsFile.Close()
				return summary, fmt.Errorf("write result document: %w", err)
			}
		}
		if _, err := writer.WriteString("    " + string(data)); err != nil {
			resultsFile.Close()
			return summary, fmt.Errorf("write result document: %w", err)
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
	if _, err := decoder.Token(); err != nil {
		resultsFile.Close()
		return summary, fmt.Errorf("decode scenario manifest: %w", err)
	}

	footer := "\n  ]\n}\n"
	if _, err := writer.WriteString(footer); err != nil {
		resultsFile.Close()
		return summary, fmt.Errorf("write result document: %w", err)
	}
	if err := writer.Flush(); err != nil {
		resultsFile.Close()
		return summary, fmt.Errorf("flush result document: %w", err)
	}
	if err := resultsFile.Close(); err != nil {
		return summary, fmt.Errorf("close result document: %w", err)
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
