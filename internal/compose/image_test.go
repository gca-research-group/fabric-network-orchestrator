package compose

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
)

type imageExecutor struct {
	commands []string
	cached   map[string]string
	failAt   int
	err      error
}

func (e *imageExecutor) ExecCommand(name string, args ...string) error {
	_, err := e.OutputCommand(name, args...)
	return err
}

func (e *imageExecutor) OutputCommand(name string, args ...string) ([]byte, error) {
	e.commands = append(e.commands, strings.Join(append([]string{name}, args...), " "))
	if len(e.commands) == e.failAt {
		return nil, e.err
	}
	if len(args) == 5 && args[0] == "image" {
		return []byte(e.cached[strings.TrimPrefix(args[4], "reference=")]), nil
	}
	return nil, nil
}

func TestPullChaincodeCompilerImages(t *testing.T) {
	const image25 = "hyperledger/fabric-ccenv:2.5"
	const image30 = "hyperledger/fabric-ccenv:3.0"
	check25 := "docker image ls --quiet --filter reference=" + image25
	check30 := "docker image ls --quiet --filter reference=" + image30
	cfg := config.Config{Organizations: []config.Organization{
		{Peers: []config.Peer{{Version: "3.0.1"}, {Version: "2.5.15"}}},
		{Peers: []config.Peer{{Version: "2.5.9"}, {Version: "3.0.2"}}},
	}}
	for _, tc := range []struct {
		name   string
		cached map[string]string
		want   []string
	}{
		{"missing", nil, []string{check25, "docker pull " + image25, check30, "docker pull " + image30}},
		{"cached", map[string]string{image25: "abc\n", image30: "def\n"}, []string{check25, check30}},
		{"mixed", map[string]string{image25: "abc\n", image30: " \n"}, []string{check25, check30, "docker pull " + image30}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec := &imageExecutor{cached: tc.cached}
			if err := PullChaincodeCompilerImages(cfg, exec); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exec.commands, tc.want) {
				t.Fatalf("want commands %v, got %v", tc.want, exec.commands)
			}
		})
	}
}

func TestPullChaincodeCompilerImagesWithoutPeers(t *testing.T) {
	for _, cfg := range []config.Config{{}, {Organizations: []config.Organization{{Name: "Org1"}}}} {
		exec := &imageExecutor{}
		if err := PullChaincodeCompilerImages(cfg, exec); err != nil {
			t.Fatal(err)
		}
		if len(exec.commands) != 0 {
			t.Fatalf("unexpected commands: %v", exec.commands)
		}
	}
}

func TestPullChaincodeCompilerImagesStopsOnFailure(t *testing.T) {
	cfg := config.Config{Organizations: []config.Organization{{Peers: []config.Peer{{Version: "2.5.15"}, {Version: "3.0.1"}}}}}
	for _, tc := range []struct {
		name   string
		failAt int
	}{
		{"check", 1},
		{"pull", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cause := errors.New("injected failure")
			exec := &imageExecutor{failAt: tc.failAt, err: cause}
			err := PullChaincodeCompilerImages(cfg, exec)
			if !errors.Is(err, cause) || !strings.Contains(err.Error(), tc.name+" compiler image hyperledger/fabric-ccenv:2.5") || strings.HasSuffix(err.Error(), "\n") {
				t.Fatalf("expected wrapped contextual error, got %v", err)
			}
			if len(exec.commands) != tc.failAt {
				t.Fatalf("continued after failure: %v", exec.commands)
			}
		})
	}
}
