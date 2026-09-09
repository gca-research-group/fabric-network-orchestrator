package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
)

// CountScenarios reads the manifest incrementally without loading the corpus.
// It checks the complete JSON document before validation creates results.json.
func CountScenarios(directory string) (int, error) {
	file, err := os.Open(filepath.Join(directory, "scenarios.json"))
	if err != nil {
		return 0, fmt.Errorf("open scenario manifest: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))
	token, err := decoder.Token()
	if err != nil {
		return 0, fmt.Errorf("decode scenario manifest: %w", err)
	}
	if token != json.Delim('[') {
		return 0, fmt.Errorf("decode scenario manifest: expected JSON array")
	}
	total := 0
	for decoder.More() {
		var scenario generator.ScenarioRules
		if err := decoder.Decode(&scenario); err != nil {
			return 0, fmt.Errorf("decode scenario manifest entry: %w", err)
		}
		if err := validateScenario(scenario); err != nil {
			return 0, fmt.Errorf("decode scenario manifest entry: %w", err)
		}
		total++
	}
	if _, err := decoder.Token(); err != nil {
		return 0, fmt.Errorf("decode scenario manifest: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return 0, fmt.Errorf("decode scenario manifest: unexpected trailing data")
		}
		return 0, fmt.Errorf("decode scenario manifest: %w", err)
	}
	return total, nil
}
