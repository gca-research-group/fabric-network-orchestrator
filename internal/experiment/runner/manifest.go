package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/parquet-go/parquet-go"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
)

// CountScenarios reads and validates the manifest incrementally without loading the corpus.
// It checks the complete Parquet document before validation creates results.parquet.
func CountScenarios(directory string) (int, error) {
	file, err := openScenarioManifest(directory)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	reader := parquet.NewGenericReader[generator.ScenarioRules](file)
	defer reader.Close()
	total := 0
	rows := make([]generator.ScenarioRules, 1024)
	for {
		n, err := reader.Read(rows)
		for _, scenario := range rows[:n] {
			if err := validateScenario(scenario); err != nil {
				return 0, fmt.Errorf("decode scenario manifest entry: %w", err)
			}
			total++
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("decode scenario manifest entry: %w", err)
		}
	}
	if total == 0 {
		return 0, fmt.Errorf("decode scenario manifest: no scenarios")
	}
	return total, nil
}

func openScenarioManifest(directory string) (*os.File, error) {
	path := filepath.Join(directory, "scenarios.parquet")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open scenario manifest: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("inspect scenario manifest: %w", err)
	}
	parquetFile, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("decode scenario manifest: %w", err)
	}
	want := parquet.SchemaOf(new(generator.ScenarioRules)).String()
	if got := parquetFile.Schema().String(); got != want {
		file.Close()
		return nil, fmt.Errorf("decode scenario manifest: incompatible schema: got %s, want %s", got, want)
	}
	return file, nil
}
