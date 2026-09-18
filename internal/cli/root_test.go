package cli

import "testing"

func TestRootCommandRegistersSupportedCommands(t *testing.T) {
	want := map[string]bool{"deploy": false, "artifacts": false, "chaincode": false, "identity": false, "image": false, "network": false}
	for _, command := range rootCmd.Commands() {
		if _, ok := want[command.Name()]; ok {
			want[command.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("root command does not register %q", name)
		}
	}
}

func TestLoadConfigRequiresPath(t *testing.T) {
	previous := configPath
	configPath = ""
	t.Cleanup(func() { configPath = previous })
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected missing configuration path to fail")
	}
}
