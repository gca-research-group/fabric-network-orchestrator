package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestSetupLogsOnlyToStdout(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temporaryDirectory := t.TempDir()

	originalStdout := os.Stdout
	originalLogger := slog.Default()
	t.Cleanup(func() {
		os.Stdout = originalStdout
		slog.SetDefault(originalLogger)
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	if err := os.Chdir(temporaryDirectory); err != nil {
		t.Fatal(err)
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer

	if err := Setup(); err != nil {
		t.Fatal(err)
	}

	slog.Info("console-only message")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "console-only message") {
		t.Fatalf("expected log message on stdout, got %q", output)
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected logger to create no files or directories, got %v", entries)
	}
}
