package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type phaseParameters struct {
	Seed             string `json:"seed,omitempty"`
	MutationCount    *int   `json:"mutationCount,omitempty"`
	Output           string `json:"output"`
	ProgressInterval int    `json:"progressInterval"`
}

type phaseMetadata struct {
	Parameters       phaseParameters `json:"parameters"`
	WorkingDirectory string          `json:"workingDirectory"`
	StartedAt        time.Time       `json:"startedAt"`
	FinishedAt       *time.Time      `json:"finishedAt,omitempty"`
	DurationSeconds  *float64        `json:"durationSeconds,omitempty"`
	Status           string          `json:"status"`
	Error            string          `json:"error,omitempty"`
}

type experimentMetadata struct {
	Generation *phaseMetadata `json:"generation,omitempty"`
	Validation *phaseMetadata `json:"validation,omitempty"`
}

func runPhase(phase string, parameters phaseParameters, stdout io.Writer, operation func() error) error {
	started := time.Now()
	fmt.Fprintf(stdout, "%s started at: %s\n", phase, started.UTC().Format(time.RFC3339Nano))
	metadata := experimentMetadata{}
	path := filepath.Join(parameters.Output, "metadata.json")
	if phase == "validation" {
		data, err := os.ReadFile(path)
		if err == nil {
			var document *experimentMetadata
			if err := json.Unmarshal(data, &document); err != nil {
				return finishPhaseOutput(phase, started, stdout, fmt.Errorf("decode experiment metadata: %w", err))
			}
			if document == nil {
				return finishPhaseOutput(phase, started, stdout, errors.New("decode experiment metadata: expected JSON object"))
			}
			metadata = *document
		} else if !errors.Is(err, os.ErrNotExist) {
			return finishPhaseOutput(phase, started, stdout, fmt.Errorf("read experiment metadata: %w", err))
		}
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return finishPhaseOutput(phase, started, stdout, fmt.Errorf("get working directory: %w", err))
	}
	record := &phaseMetadata{Parameters: parameters, WorkingDirectory: workingDirectory, StartedAt: started.UTC(), Status: "running"}
	if phase == "generation" {
		metadata.Generation = record
	} else {
		metadata.Validation = record
	}
	if err := writeMetadata(path, metadata); err != nil {
		return finishPhaseOutput(phase, started, stdout, err)
	}
	operationErr := operation()
	finished := time.Now()
	finishedUTC := finished.UTC()
	duration := finished.Sub(started).Seconds()
	record.FinishedAt, record.DurationSeconds = &finishedUTC, &duration
	record.Status = "succeeded"
	if operationErr != nil {
		record.Status, record.Error = "failed", operationErr.Error()
	}
	writeErr := writeMetadata(path, metadata)
	fmt.Fprintf(stdout, "%s finished at: %s; elapsed: %s\n", phase, finishedUTC.Format(time.RFC3339Nano), finished.Sub(started))
	return errors.Join(operationErr, writeErr)
}

func finishPhaseOutput(phase string, started time.Time, stdout io.Writer, err error) error {
	finished := time.Now()
	fmt.Fprintf(stdout, "%s finished at: %s; elapsed: %s\n", phase, finished.UTC().Format(time.RFC3339Nano), finished.Sub(started))
	return err
}

func writeMetadata(path string, metadata experimentMetadata) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create metadata directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".metadata-*.json")
	if err != nil {
		return fmt.Errorf("create experiment metadata: %w", err)
	}
	defer os.Remove(file.Name())
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(metadata)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write experiment metadata: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace experiment metadata: %w", err)
	}
	return nil
}
