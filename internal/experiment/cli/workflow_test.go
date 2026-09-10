package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parquet-go/parquet-go"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/runner"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/seed"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
)

func TestCommandHelpAndErrors(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}, {"generate", "--help"}, {"validate", "--help"}, {"seed", "--help"}, {"seed", "generate", "--help"}} {
		var output strings.Builder
		if err := run(args, &output); err != nil || !strings.Contains(output.String(), "Usage") {
			t.Fatalf("help %v: %v, %s", args, err, output.String())
		}
	}
	for _, args := range [][]string{nil, {"--seed", "seed.yaml"}, {"unknown"}, {"seed"}, {"seed", "unknown"}, {"generate-seed"}, {"validate", "--seed", "x"}, {"validate", "--mutation-count", "1"}, {"generate", "--seed", "x", "extra"}, {"seed", "generate", "extra"}, {"validate", "extra"}, {"seed", "generate", "--unknown"}} {
		if err := run(args, &strings.Builder{}); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
}

func TestSeedGeneration(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := run([]string{"seed", "generate"}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, "seed.yaml"); string(got) != seed.YAML {
		t.Fatal("seed differs from bundled template")
	}
	if _, err := config.LoadConfigFromPath("seed.yaml"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, "seed.yaml", []byte("preserve me"))
	if err := run([]string{"seed", "generate"}, &strings.Builder{}); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("overwrite protection: %v", err)
	}
	if string(readFile(t, "seed.yaml")) != "preserve me" {
		t.Fatal("existing seed changed")
	}
	if err := run([]string{"seed", "generate", "--force"}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	if string(readFile(t, "seed.yaml")) != seed.YAML {
		t.Fatal("force did not replace seed")
	}
	if err := run([]string{"seed", "generate", "--output", "nested/custom.yml"}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	if string(readFile(t, "nested/custom.yml")) != seed.YAML {
		t.Fatal("custom seed differs")
	}
}

func TestSeparateWorkflow(t *testing.T) {
	directory := t.TempDir()
	seedPath := filepath.Join(directory, "seed.yaml")
	if err := run([]string{"seed", "generate", "--output", seedPath}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	args := []string{"generate", "--seed", seedPath, "--mutation-count", "1", "--output", directory, "--progress-interval", "2"}
	var output strings.Builder
	if err := run(args, &output); err != nil {
		t.Fatal(err)
	}
	resultsPath := filepath.Join(directory, "results.parquet")
	if _, err := os.Stat(resultsPath); !os.IsNotExist(err) {
		t.Fatalf("generation wrote results: %v", err)
	}
	if strings.Contains(output.String(), "validation") {
		t.Fatal("generation ran validation")
	}
	if _, err := os.Stat(filepath.Join(directory, "config")); !os.IsNotExist(err) {
		t.Fatalf("generation created a config directory: %v", err)
	}
	if !strings.Contains(output.String(), "generation progress: 2/") || strings.Contains(output.String(), "generation progress: 1/") {
		t.Fatalf("generation ignored progress interval: %s", output.String())
	}
	writeFile(t, resultsPath, []byte("existing results"))
	if err := run(args, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	if string(readFile(t, resultsPath)) != "existing results" {
		t.Fatal("generation changed results")
	}
	if err := os.Remove(seedPath); err != nil {
		t.Fatal(err)
	}
	before := map[string][]byte{}
	if err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && path != resultsPath && entry.Name() != "metadata.json" {
			before[path] = readFile(t, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := run([]string{"validate", "--output", directory, "--progress-interval", "3"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "validation progress: 3/") || strings.Contains(output.String(), "validation progress: 2/") {
		t.Fatalf("validation ignored progress interval: %s", output.String())
	}
	for path, contents := range before {
		if !bytes.Equal(contents, readFile(t, path)) {
			t.Fatalf("validation modified %s", path)
		}
	}
	results := readResults(t, resultsPath)
	metadata := readMetadata(t, directory)
	if metadata.Validation.Total == nil || metadata.Validation.Passed == nil || *metadata.Validation.Total == 0 || *metadata.Validation.Passed != *metadata.Validation.Total {
		t.Fatalf("validation metadata: %+v", metadata.Validation)
	}
	if len(results) != *metadata.Validation.Total {
		t.Fatalf("results: %d, total: %d", len(results), *metadata.Validation.Total)
	}
	if !strings.Contains(output.String(), "(100.0%)") {
		t.Fatal("missing final progress")
	}
}

func TestValidationErrors(t *testing.T) {
	for _, manifest := range []string{"", "{}", "[", "[{}", "[1]", "[] {}", "[] garbage"} {
		t.Run(manifest, func(t *testing.T) {
			directory := t.TempDir()
			if manifest != "" {
				writeFile(t, filepath.Join(directory, "scenarios.parquet"), []byte(manifest))
			}
			if err := run([]string{"validate", "--output", directory}, &strings.Builder{}); err == nil {
				t.Fatal("expected manifest error")
			}
			if _, err := os.Stat(filepath.Join(directory, "results.parquet")); !os.IsNotExist(err) {
				t.Fatal("invalid manifest wrote results")
			}
		})
	}
}

func TestInvalidManifestPreservesExistingResults(t *testing.T) {
	directory := t.TempDir()
	resultsPath := filepath.Join(directory, "results.parquet")
	existing := []byte("existing results")
	writeFile(t, resultsPath, existing)
	writeFile(t, filepath.Join(directory, "scenarios.parquet"), []byte("invalid parquet"))

	if err := run([]string{"validate", "--output", directory}, &strings.Builder{}); err == nil {
		t.Fatal("expected manifest error")
	}
	if got := readFile(t, resultsPath); !bytes.Equal(got, existing) {
		t.Fatal("invalid manifest replaced existing results")
	}
}

func scenarioMutations(rules ...validate.RuleID) []generator.ScenarioMutation {
	result := make([]generator.ScenarioMutation, 0, len(rules))
	for _, rule := range rules {
		result = append(result, generator.ScenarioMutation{Rule: rule, OperatorIndex: 0})
	}
	return result
}

type parquetResult struct {
	Scenario string            `parquet:"scenario"`
	Expected []validate.RuleID `parquet:"expectedRules,list"`
	Actual   []validate.RuleID `parquet:"actualRules,list"`
	Missing  []validate.RuleID `parquet:"missingRules,list"`
	Status   runner.Status     `parquet:"status"`
	Error    *string           `parquet:"error,optional"`
}

func writeScenarioManifest(t *testing.T, directory string, scenarios []generator.ScenarioRules) {
	t.Helper()
	file, err := os.Create(filepath.Join(directory, "scenarios.parquet"))
	if err != nil {
		t.Fatal(err)
	}
	writer := parquet.NewGenericWriter[generator.ScenarioRules](file)
	writer.SetKeyValueMetadata(generator.SeedYAMLMetadataKey, seed.YAML)
	if _, err := writer.Write(scenarios); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func readResults(t *testing.T, path string) []parquetResult {
	t.Helper()
	rows, err := parquet.ReadFile[parquetResult](path)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
