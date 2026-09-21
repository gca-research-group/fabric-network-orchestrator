package validate

import (
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
	"testing"
)

func TestEmptyPeerNameFn(t *testing.T) {
	assertValidationError(t, EmptyPeerNameFn(spec.Peer{}, 3, "Org1"), RulePeerNameRequired, "Empty Peer Name", "name of the peer index 3 of the organization Org1 is undefined")
	assertNoError(t, EmptyPeerNameFn(spec.Peer{Name: "peer0"}, 0, "Org1"))
}

func TestEmptyPeerSubdomainFn(t *testing.T) {
	peer := spec.Peer{Name: "peer0"}
	assertValidationError(t, EmptyPeerSubdomainFn(peer, "Org1"), RulePeerSubdomainRequired, "Empty Peer Subdomain", "subdomain of the peer peer0 of the organization Org1 is undefined")
	peer.Subdomain = "peer0"
	assertNoError(t, EmptyPeerSubdomainFn(peer, "Org1"))
}

func TestInvalidPeerVersionFn(t *testing.T) {
	tests := []struct {
		name                  string
		version               string
		channelCapability     string
		applicationCapability string
		minimum               string
	}{
		{name: "application capability is stricter", version: "2.4.0", channelCapability: "V2_0", applicationCapability: "V2_5", minimum: "2.5.0"},
		{name: "channel capability is stricter", version: "2.5.0", channelCapability: "V3_0", applicationCapability: "V2_0", minimum: "3.0.0"},
		{name: "equal capabilities", version: "2.4.0", channelCapability: "V2_5", applicationCapability: "V2_5", minimum: "2.5.0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertValidationError(t, InvalidPeerVersionFn(spec.Peer{Version: test.version}, "Org1", test.channelCapability, test.applicationCapability), RulePeerVersionInvalid, "Invalid Peer Version", "peer version of org Org1 invalid: version "+test.version+" is lower than required "+test.minimum)
		})
	}

	for _, version := range []string{"", "2.5.0", "3.0.0"} {
		t.Run("valid "+version, func(t *testing.T) {
			assertNoError(t, InvalidPeerVersionFn(spec.Peer{Version: version}, "Org1", "V2_0", "V2_5"))
		})
	}
}
