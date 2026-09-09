package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	resultsPath := filepath.Join(directory, "results.json")
	if _, err := os.Stat(resultsPath); !os.IsNotExist(err) {
		t.Fatalf("generation wrote results: %v", err)
	}
	if strings.Contains(output.String(), "validation") {
		t.Fatal("generation ran validation")
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
	var summary runner.Summary
	if err := json.Unmarshal(readFile(t, resultsPath), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Total == 0 || summary.Passed != summary.Total {
		t.Fatalf("summary: %+v", summary)
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
				writeFile(t, filepath.Join(directory, "scenarios.json"), []byte(manifest))
			}
			if err := run([]string{"validate", "--output", directory}, &strings.Builder{}); err == nil {
				t.Fatal("expected manifest error")
			}
			if _, err := os.Stat(filepath.Join(directory, "results.json")); !os.IsNotExist(err) {
				t.Fatal("invalid manifest wrote results")
			}
		})
	}
}

func TestValidationUnsuccessfulScenarios(t *testing.T) {
	for _, status := range []string{"partial", "failed", "missing"} {
		t.Run(status, func(t *testing.T) {
			directory := t.TempDir()
			rules := []validate.RuleID{validate.RuleChannelNameInvalid}
			if status == "partial" {
				rules = append(rules, validate.RuleOrganizationsRequired)
			}
			manifest, err := json.Marshal([]generator.ScenarioRules{{Scenario: "000001", Rules: rules}})
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(directory, "scenarios.json"), manifest)
			if status != "missing" {
				if err := os.Mkdir(filepath.Join(directory, "config"), 0755); err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepath.Join(directory, "config", "000001.yaml"), []byte("output: output/example\norganizations: []\n"))
			}
			if err := run([]string{"validate", "--output", directory}, &strings.Builder{}); err == nil || !strings.Contains(err.Error(), "expected validation rules") {
				t.Fatalf("expected failed verification: %v", err)
			}
			var document struct {
				Results []runner.Result `json:"results"`
			}
			if err := json.Unmarshal(readFile(t, filepath.Join(directory, "results.json")), &document); err != nil {
				t.Fatal(err)
			}
			want := runner.StatusFailed
			if status == "partial" {
				want = runner.StatusPartial
			}
			if len(document.Results) != 1 || document.Results[0].Status != want {
				t.Fatalf("results: %+v", document)
			}
		})
	}
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
