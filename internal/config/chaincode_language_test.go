package config

import (
	"strings"
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
	"gopkg.in/yaml.v3"
)

func TestChaincodeLanguageDefault(t *testing.T) {
	cfg := getValidBaseConfig()
	cfg.Channels[0].Chaincodes = []Chaincode{{Name: "Asset", Path: "asset", Version: "1.0"}}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfigFromYAML(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Channels[0].Chaincodes[0].Language.Name; got != spec.LanguageGo {
		t.Fatalf("expected Go default, got %q", got)
	}
}

func TestUnsupportedChaincodeLanguageOnLoad(t *testing.T) {
	cfg := getValidBaseConfig()
	cfg.Channels[0].Chaincodes = []Chaincode{{Name: "Asset", Path: "asset", Version: "1.0", Language: Language{Name: "python"}}}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = LoadConfigFromYAML(data)
	if err == nil || !strings.Contains(err.Error(), "chaincode.language.unsupported") {
		t.Fatalf("expected language validation error, got %v", err)
	}
}

func TestLegacyLanguageVersionIsIgnored(t *testing.T) {
	cfg := getValidBaseConfig()
	cfg.Channels[0].Chaincodes = []Chaincode{{Name: "Asset", Path: "asset", Version: "1.0", Language: Language{Name: spec.LanguageGo}}}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	legacy := strings.Replace(string(data), "name: golang", "name: golang\n            version: \"1.26\"", 1)
	if legacy == string(data) {
		t.Fatal("could not add legacy language.version field")
	}
	loaded, err := LoadConfigFromYAML([]byte(legacy))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Channels[0].Chaincodes[0].Language.Name != spec.LanguageGo {
		t.Fatalf("unexpected language: %+v", loaded.Channels[0].Chaincodes[0].Language)
	}
	current, err := yaml.Marshal(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(current), "version: \"1.26\"") {
		t.Fatal("legacy language.version survived loading")
	}
}
