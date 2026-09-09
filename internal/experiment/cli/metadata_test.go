package cli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/seed"
)

func readMetadata(t *testing.T, directory string) experimentMetadata {
	t.Helper()
	var metadata experimentMetadata
	if err := json.Unmarshal(readFile(t, filepath.Join(directory, "metadata.json")), &metadata); err != nil {
		t.Fatal(err)
	}
	return metadata
}

func checkFinished(t *testing.T, record *phaseMetadata, status string) {
	t.Helper()
	if record == nil || record.Status != status || record.StartedAt.IsZero() || record.FinishedAt == nil || record.DurationSeconds == nil {
		t.Fatalf("incomplete metadata: %+v", record)
	}
	if record.FinishedAt.Before(record.StartedAt) || *record.DurationSeconds < 0 {
		t.Fatalf("invalid timing: %+v", record)
	}
	_, offset := record.StartedAt.Zone()
	if offset != 0 {
		t.Fatal("timestamp is not UTC")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if record.WorkingDirectory != workingDirectory {
		t.Fatalf("working directory: %s", record.WorkingDirectory)
	}
}

func TestMetadataLifecycle(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, "seed.yaml", []byte(seed.YAML))
	args := []string{"generate", "--seed", "seed.yaml", "--mutation-count", "1", "--output", "corpus", "--progress-interval", "7"}
	var output strings.Builder
	if err := run(args, &output); err != nil {
		t.Fatal(err)
	}
	generated := readMetadata(t, "corpus")
	checkFinished(t, generated.Generation, "succeeded")
	p := generated.Generation.Parameters
	if p.Seed != "seed.yaml" || p.MutationCount == nil || *p.MutationCount != 1 || p.Output != "corpus" || p.ProgressInterval != 7 {
		t.Fatalf("parameters: %+v", p)
	}
	if generated.Validation != nil {
		t.Fatal("generation included validation")
	}
	if err := run([]string{"validate", "--output", "corpus"}, &output); err != nil {
		t.Fatal(err)
	}
	validated := readMetadata(t, "corpus")
	checkFinished(t, validated.Validation, "succeeded")
	if !reflect.DeepEqual(generated.Generation, validated.Generation) {
		t.Fatal("validation changed generation metadata")
	}
	if validated.Validation.Parameters.ProgressInterval != 1000 || validated.Validation.Parameters.Output != "corpus" || validated.Validation.Parameters.MutationCount != nil {
		t.Fatal("validation parameters incorrect")
	}
	for _, phase := range []string{"generation", "validation"} {
		if !strings.Contains(output.String(), phase+" started at:") || !strings.Contains(output.String(), phase+" finished at:") || !strings.Contains(output.String(), "elapsed:") {
			t.Fatalf("missing timing: %s", output.String())
		}
	}
	if err := run(args, io.Discard); err != nil {
		t.Fatal(err)
	}
	if readMetadata(t, "corpus").Validation != nil {
		t.Fatal("regeneration retained stale validation")
	}
}

func TestMetadataRecordsFailuresAndDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	err := run([]string{"generate", "--seed", "missing.yaml"}, io.Discard)
	if err == nil {
		t.Fatal("expected unreadable seed")
	}
	generation := readMetadata(t, "output").Generation
	checkFinished(t, generation, "failed")
	if generation.Error != err.Error() || generation.Parameters.MutationCount == nil || *generation.Parameters.MutationCount != 3 || generation.Parameters.ProgressInterval != 1000 || generation.Parameters.Output != "output" {
		t.Fatalf("failed generation: %+v", generation)
	}
	for _, manifest := range []string{"[", `[{"scenario":"missing","rules":["unknown"]}]`} {
		directory := t.TempDir()
		writeFile(t, filepath.Join(directory, "scenarios.json"), []byte(manifest))
		err := run([]string{"validate", "--output", directory}, io.Discard)
		if err == nil {
			t.Fatal("expected validation error")
		}
		metadata := readMetadata(t, directory)
		checkFinished(t, metadata.Validation, "failed")
		if metadata.Generation != nil || metadata.Validation.Error != err.Error() {
			t.Fatalf("failed validation: %+v", metadata)
		}
	}
}

func TestMetadataRunningRecord(t *testing.T) {
	directory := t.TempDir()
	if err := runPhase("validation", phaseParameters{Output: directory, ProgressInterval: 10}, io.Discard, func() error {
		record := readMetadata(t, directory).Validation
		if record.Status != "running" || record.FinishedAt != nil || record.DurationSeconds != nil {
			t.Fatalf("running record: %+v", record)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	checkFinished(t, readMetadata(t, directory).Validation, "succeeded")
}

func TestMalformedMetadataPreserved(t *testing.T) {
	for _, contents := range []string{"{", "null", "[]", `{"generation":3}`, "{} {}"} {
		directory := t.TempDir()
		path := filepath.Join(directory, "metadata.json")
		writeFile(t, path, []byte(contents))
		if err := run([]string{"validate", "--output", directory}, io.Discard); err == nil || !strings.Contains(err.Error(), "decode experiment metadata") {
			t.Fatalf("metadata %q: %v", contents, err)
		}
		if string(readFile(t, path)) != contents {
			t.Fatal("malformed metadata replaced")
		}
	}
}

func TestInvalidArgumentsDoNotCreateMetadata(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{{"generate"}, {"generate", "--seed="}, {"generate", "--seed", "seed.yaml", "--progress-interval", "0"}, {"validate", "--progress-interval", "-1"}, {"generate", "--seed", "seed.yaml", "--mutation-count", "invalid"}} {
		if err := run(args, io.Discard); err == nil {
			t.Fatalf("expected error: %v", args)
		}
		if _, err := os.Stat("output"); !os.IsNotExist(err) {
			t.Fatalf("invalid arguments created output: %v", err)
		}
	}
}

func TestMetadataWriteFailures(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "metadata.json")
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	called := false
	err := runPhase("generation", phaseParameters{Output: directory}, io.Discard, func() error { called = true; return nil })
	if err == nil || called {
		t.Fatalf("operation ran despite initial write failure: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	original := errors.New("operation failed")
	err = runPhase("generation", phaseParameters{Output: directory}, io.Discard, func() error {
		// A directory at the destination forces the final rename to fail on all platforms.
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatal(err)
		}
		return original
	})
	if !errors.Is(err, original) || !strings.Contains(err.Error(), "replace experiment metadata") {
		t.Fatalf("errors not preserved: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(directory, ".metadata-*.json"))
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary metadata files left: %v, %v", files, err)
	}
}
