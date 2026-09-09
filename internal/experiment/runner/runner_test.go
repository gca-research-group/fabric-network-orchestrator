package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/format"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
)

func TestRunChecksEveryScenario(t *testing.T) {
	outputDirectory := t.TempDir()
	configDirectory := filepath.Join(outputDirectory, "config")
	if err := os.Mkdir(configDirectory, 0755); err != nil {
		t.Fatal(err)
	}

	scenarios := []generator.ScenarioRules{
		{Scenario: "000001", Mutations: mutations(validate.RuleOrganizationsRequired, validate.RuleOrdererTopologyRequired)},
		{Scenario: "000002", Mutations: mutations(validate.RuleOrganizationsRequired, validate.RuleApplicationCapabilityUnsupported)},
		{Scenario: "000003", Mutations: mutations(validate.RuleChannelNameInvalid)},
	}
	writeScenario(t, configDirectory, "000001", "output: output/example\norganizations: []\n")
	writeScenario(t, configDirectory, "000002", "output: output/example\ncapabilities:\n  channel: V2_0\n  application: V2_5\n  orderer: V2_0\norganizations: []\n")
	writeScenario(t, configDirectory, "000003", "output: output/example\norganizations: []\n")
	if err := parquet.WriteFile(filepath.Join(outputDirectory, "scenarios.parquet"), scenarios); err != nil {
		t.Fatal(err)
	}

	var progress []int
	summary, err := RunDirectory(outputDirectory, func(completed int) {
		progress = append(progress, completed)
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 3 || summary.Passed != 1 || summary.Partial != 1 || summary.Failed != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(progress) != 3 || progress[0] != 1 || progress[1] != 2 || progress[2] != 3 {
		t.Fatalf("unexpected progress updates: %v", progress)
	}
	rows, err := parquet.ReadFile[resultRow](filepath.Join(outputDirectory, "results.parquet"))
	if err != nil {
		t.Fatalf("decode results: %v", err)
	}
	if rows[0].Status != StatusPassed || rows[1].Status != StatusPartial || rows[2].Status != StatusFailed {
		t.Fatalf("unexpected result states: %+v", rows)
	}
	if len(rows[1].Missing) != 1 || rows[1].Missing[0] != validate.RuleApplicationCapabilityUnsupported {
		t.Fatalf("unexpected missing rules: %+v", rows[1].Missing)
	}
	assertZstdResults(t, filepath.Join(outputDirectory, "results.parquet"))
}

func TestCountScenariosIgnoresLegacyJSONManifest(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "scenarios.json"), []byte(`[{"scenario":"000001","rules":["organizations.required"]}]`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := CountScenarios(directory); err == nil {
		t.Fatal("expected legacy manifest to be rejected")
	}
}

func TestCountScenariosRejectsEmptyAndIncompatibleParquet(t *testing.T) {
	for name, rows := range map[string]any{
		"empty":        []generator.ScenarioRules{},
		"incompatible": []struct{ Other string }{{Other: "value"}},
	} {
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "scenarios.parquet")
			var err error
			switch value := rows.(type) {
			case []generator.ScenarioRules:
				err = parquet.WriteFile(path, value)
			case []struct{ Other string }:
				err = parquet.WriteFile(path, value)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := CountScenarios(directory); err == nil {
				t.Fatal("expected manifest to be rejected")
			}
		})
	}
}

func assertZstdResults(t *testing.T, path string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	document, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range document.Metadata().RowGroups {
		for _, column := range group.Columns {
			if column.MetaData.Codec != format.Zstd {
				t.Fatalf("column %v uses %s compression", column.MetaData.PathInSchema, column.MetaData.Codec)
			}
		}
	}
}

func mutations(rules ...validate.RuleID) []generator.ScenarioMutation {
	result := make([]generator.ScenarioMutation, 0, len(rules))
	for _, rule := range rules {
		result = append(result, generator.ScenarioMutation{Rule: rule, OperatorIndex: 0})
	}
	return result
}

func writeScenario(t *testing.T, directory, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name+".yaml"), []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}
