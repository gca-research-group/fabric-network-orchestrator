package application

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
)

type fakeExecutor struct {
	commands []string
	output   []byte
	failAt   int
}

func (f *fakeExecutor) ExecCommand(name string, args ...string) error {
	f.commands = append(f.commands, strings.Join(append([]string{name}, args...), " "))
	if f.failAt == len(f.commands) {
		return errors.New("injected command failure")
	}
	return nil
}

func (f *fakeExecutor) OutputCommand(name string, args ...string) ([]byte, error) {
	f.commands = append(f.commands, strings.Join(append([]string{name}, args...), " "))
	if f.failAt == len(f.commands) {
		return nil, errors.New("injected command failure")
	}
	return f.output, nil
}

func TestGenerateArtifactsRejectsNonEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := NewWorkflows(&fakeExecutor{}).GenerateArtifacts(&config.Config{Output: dir})
	if err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("expected non-empty directory error, got %v", err)
	}
}

func TestDeployPreservesExistingArtifactsAndStopsBeforeNetwork(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	exec := &fakeExecutor{}
	cfg := artifactTestConfig(dir)
	err := NewWorkflows(exec).Deploy(&cfg)
	if err == nil || !strings.Contains(err.Error(), "stage generate artifacts:") {
		t.Fatalf("expected artifact stage failure, got %v", err)
	}
	if len(exec.commands) != 0 {
		t.Fatalf("deployment continued after artifact failure: %v", exec.commands)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "data" {
		t.Fatalf("existing artifacts changed: %q, %v", got, err)
	}
}

func TestDeployGeneratesArtifactsBeforeNetworkAndStopsOnNetworkFailure(t *testing.T) {
	cfg := artifactTestConfig(t.TempDir())
	exec := &fakeExecutor{failAt: 1}
	err := NewWorkflows(exec).Deploy(&cfg)
	if err == nil || !strings.Contains(err.Error(), "stage deploy network:") {
		t.Fatalf("expected network stage failure, got %v", err)
	}
	for _, name := range []string{"configtx.yml", "network.yml"} {
		if _, err := os.Stat(filepath.Join(cfg.Output, name)); err != nil {
			t.Fatalf("artifact %s was not preserved: %v", name, err)
		}
	}
	if len(exec.commands) != 1 || !strings.HasPrefix(exec.commands[0], "docker ") {
		t.Fatalf("unexpected commands after network failure: %v", exec.commands)
	}
}

func TestDeployChaincodesPreparesCompilerBeforePublication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		output []byte
		want   []string
	}{
		{"missing", nil, []string{
			"docker image ls --quiet --filter reference=hyperledger/fabric-ccenv:2.5",
			"docker pull hyperledger/fabric-ccenv:2.5",
		}},
		{"cached", []byte("image-id\n"), []string{
			"docker image ls --quiet --filter reference=hyperledger/fabric-ccenv:2.5",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec := &fakeExecutor{output: tc.output, failAt: len(tc.want) + 1}
			cfg := artifactTestConfig("artifacts")
			cfg.Channels[0].Chaincodes = []config.Chaincode{{Name: "asset", Version: "1.0"}}
			err := NewWorkflows(exec).DeployChaincodes(&cfg)
			if err == nil || !strings.Contains(err.Error(), "Package the Chaincodes") {
				t.Fatalf("expected publication to start after preparation, got %v", err)
			}
			if len(exec.commands) != len(tc.want)+1 || !reflect.DeepEqual(exec.commands[:len(tc.want)], tc.want) || !strings.HasPrefix(exec.commands[len(tc.want)], "docker exec ") {
				t.Fatalf("unexpected preparation/publication order: %v", exec.commands)
			}
		})
	}
}

func TestDeployChaincodesStopsOnCompilerPreparationFailure(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		exec := &fakeExecutor{failAt: failAt}
		cfg := artifactTestConfig("artifacts")
		cfg.Channels[0].Chaincodes = []config.Chaincode{{Name: "asset", Version: "1.0"}}
		err := NewWorkflows(exec).DeployChaincodes(&cfg)
		if err == nil || !strings.Contains(err.Error(), "deploy chaincodes: prepare compiler images:") {
			t.Fatalf("expected compiler preparation failure, got %v", err)
		}
		if len(exec.commands) != failAt {
			t.Fatalf("publication continued after preparation failure: %v", exec.commands)
		}
	}
}

func TestStopNetworkUsesInjectedExecutor(t *testing.T) {
	exec := &fakeExecutor{output: []byte("peer0.example.org\norderer.example.org\n")}
	err := NewWorkflows(exec).StopNetwork(&config.Config{Network: "example"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"docker network inspect example --format {{ range .Containers }}{{ .Name }}{{ \"\\n\" }}{{ end }}",
		"docker rm -f peer0.example.org",
		"docker rm -f orderer.example.org",
	}
	if !reflect.DeepEqual(exec.commands, want) {
		t.Fatalf("commands mismatch\nwant: %#v\n got: %#v", want, exec.commands)
	}
}

func TestStartNetworkStopsAtFirstCommandFailure(t *testing.T) {
	exec := &fakeExecutor{failAt: 2}
	cfg := config.Config{
		Output: "artifacts",
		Organizations: []config.Organization{{
			Name: "Org1", Domain: "org1.example.org",
			Orderers: []config.Orderer{{Name: "Orderer", Subdomain: "orderer"}},
			Peers:    []config.Peer{{Name: "Peer", Subdomain: "peer0"}},
		}},
	}
	err := NewWorkflows(exec).StartNetwork(&cfg)
	if err == nil || !strings.Contains(err.Error(), "Start Orderers") {
		t.Fatalf("expected contextual orderer failure, got %v", err)
	}
	if len(exec.commands) != 2 {
		t.Fatalf("expected workflow to stop after two commands, got %d", len(exec.commands))
	}
}

func TestGenerateArtifactsMatchesGoldenFiles(t *testing.T) {
	output := t.TempDir()
	cfg := artifactTestConfig(output)
	if err := NewWorkflows(&fakeExecutor{}).GenerateArtifacts(&cfg); err != nil {
		t.Fatal(err)
	}

	files := []string{
		"configtx.yml",
		"network.yml",
		filepath.Join("org1.example.org", "ca.org1.example.org.yml"),
		filepath.Join("org1.example.org", "peer0.org1.example.org.yml"),
	}
	for _, name := range files {
		got, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		golden := filepath.Join("testdata", "artifacts", name)
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(golden, got, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("generated artifact differs from golden file: %s", name)
		}
	}
}

func artifactTestConfig(output string) config.Config {
	return config.Config{
		Output: output, Network: "example",
		Capabilities: config.Capabilities{Channel: "V2_0", Orderer: "V2_0", Application: "V2_0"},
		Organizations: []config.Organization{{
			Name: "Org1", Domain: "org1.example.org", Bootstrap: true,
			CertificateAuthority: config.CertificateAuthority{ExposePort: 7054, Version: "latest"},
			Orderers:             []config.Orderer{{Name: "Orderer", Subdomain: "orderer", Port: 7050, ExposePort: 7050, Version: "2.5.15"}},
			Peers:                []config.Peer{{Name: "Peer0", Subdomain: "peer0", Port: 7051, ExposePort: 7051, Version: "2.5.15", IsAnchor: true}},
		}},
		Profiles: []config.Profile{{Name: "Default", Organizations: []string{"Org1"}, Consensus: config.Consensus{Type: "etcdraft"}}},
		Channels: []config.Channel{{Name: "defaultchannel", Profile: "Default"}},
	}
}
