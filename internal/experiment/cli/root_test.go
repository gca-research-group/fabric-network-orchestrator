package cli

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/generator"
	"github.com/spf13/cobra"
)

func run(args []string, stdout io.Writer) error {
	command := newRootCommand()
	command.SetArgs(args)
	command.SetOut(stdout)
	command.SetErr(stdout)
	return command.Execute()
}

func TestCommandFlags(t *testing.T) {
	for _, test := range []struct {
		path     []string
		defaults map[string]string
	}{
		{[]string{"generate"}, map[string]string{"seed": "", "mutation-count": strconv.Itoa(generator.DefaultMutationCount), "output": "output", "progress-interval": "1000"}},
		{[]string{"validate"}, map[string]string{"output": "output", "progress-interval": "1000"}},
		{[]string{"seed", "generate"}, map[string]string{"output": "seed.yaml", "force": "false"}},
	} {
		command, _, err := newRootCommand().Find(test.path)
		if err != nil {
			t.Fatal(err)
		}
		for name, want := range test.defaults {
			flag := command.Flags().Lookup(name)
			if flag == nil || flag.DefValue != want {
				t.Fatalf("%v flag %s: %v, want %s", test.path, name, flag, want)
			}
		}
	}
}

func TestGenerateFlagOverridesAndIsolation(t *testing.T) {
	command := newRootCommand()
	generate, _, err := command.Find([]string{"generate"})
	if err != nil {
		t.Fatal(err)
	}
	generate.RunE = func(cmd *cobra.Command, args []string) error {
		for name, want := range map[string]string{"seed": "custom.yaml", "mutation-count": "2", "output": "custom-output"} {
			if got := cmd.Flags().Lookup(name).Value.String(); got != want {
				t.Errorf("%s = %s, want %s", name, got, want)
			}
		}
		return nil
	}
	command.SetArgs([]string{"generate", "--seed", "custom.yaml", "--mutation-count", "2", "--output", "custom-output"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"generate"}, io.Discard); err == nil {
		t.Fatal("seed flag leaked into next invocation")
	}
}

func TestGenerateErrors(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, test := range []struct {
		args    []string
		message string
	}{
		{[]string{"generate"}, "seed"},
		{[]string{"generate", "--seed="}, "--seed"},
		{[]string{"generate", "--seed", "missing-seed.yml"}, "read seed YAML"},
		{[]string{"generate", "--seed", "seed.yml", "--mutation-count", "invalid"}, "invalid"},
		{[]string{"--seed", "seed.yml"}, migrationGuidance},
	} {
		if err := run(test.args, io.Discard); err == nil || !strings.Contains(err.Error(), test.message) {
			t.Fatalf("%v: %v, want %s", test.args, err, test.message)
		}
	}
}
