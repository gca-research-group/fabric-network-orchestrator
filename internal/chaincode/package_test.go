package chaincode

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

type packageExecutor struct{ commands [][]string }

func (e *packageExecutor) ExecCommand(name string, args ...string) error {
	return nil
}
func (e *packageExecutor) OutputCommand(name string, args ...string) ([]byte, error) {
	e.commands = append(e.commands, append([]string{name}, args...))
	return nil, nil
}

func TestPackageLanguages(t *testing.T) {
	for _, tc := range []struct {
		language string
		steps    int
	}{
		{spec.LanguageGo, 2}, {spec.LanguageJava, 1}, {spec.LanguageNode, 1},
	} {
		t.Run(tc.language, func(t *testing.T) {
			e := &packageExecutor{}
			cfg := &config.Config{Organizations: []config.Organization{{Name: "Org1"}}, Channels: []config.Channel{{Chaincodes: []config.Chaincode{{Name: "Asset", Path: "samples/asset", Version: "1.0", Language: config.Language{Name: tc.language}}}}}}
			if err := NewChaincode(cfg, e).Package(); err != nil {
				t.Fatal(err)
			}
			if len(e.commands) != tc.steps {
				t.Fatalf("got %d commands: %v", len(e.commands), e.commands)
			}
			last := e.commands[len(e.commands)-1]
			if !reflect.DeepEqual(last[len(last)-6:], []string{"--path", "/chaincodes/asset", "--lang", tc.language, "--label", "Asset_1.0"}) {
				t.Fatalf("unexpected package command: %v", last)
			}
			if tc.steps == 2 && !strings.Contains(strings.Join(e.commands[0], " "), "go mod tidy") {
				t.Fatalf("missing Go preparation: %v", e.commands[0])
			}
		})
	}
}
