package validate

import (
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
	"testing"
)

func TestEmptyOrdererNameFn(t *testing.T) {
	assertValidationError(t, EmptyOrdererNameFn(spec.Orderer{}, 2, "Org1"), RuleOrdererNameRequired, "Empty Orderer Name", "name of the orderer index 2 of the organization Org1 is undefined")
	assertNoError(t, EmptyOrdererNameFn(spec.Orderer{Name: "orderer0"}, 0, "Org1"))
}

func TestEmptyOrdererSubdomainFn(t *testing.T) {
	orderer := spec.Orderer{Name: "orderer0"}
	assertValidationError(t, EmptyOrdererSubdomainFn(orderer, "Org1"), RuleOrdererSubdomainRequired, "Empty Orderer Subdomain", "subdomain of the orderer orderer0 of the organization Org1 is undefined")
	orderer.Subdomain = "orderer0"
	assertNoError(t, EmptyOrdererSubdomainFn(orderer, "Org1"))
}

func TestInvalidOrdererVersionFn(t *testing.T) {
	tests := []struct {
		name              string
		version           string
		channelCapability string
		ordererCapability string
		minimum           string
	}{
		{name: "orderer capability is stricter", version: "2.4.0", channelCapability: "V2_0", ordererCapability: "V2_5", minimum: "2.5.0"},
		{name: "channel capability is stricter", version: "2.5.0", channelCapability: "V3_0", ordererCapability: "V2_0", minimum: "3.0.0"},
		{name: "equal capabilities", version: "2.4.0", channelCapability: "V2_5", ordererCapability: "V2_5", minimum: "2.5.0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertValidationError(t, InvalidOrdererVersionFn(spec.Orderer{Version: test.version}, "Org1", test.channelCapability, test.ordererCapability), RuleOrdererVersionInvalid, "Invalid Orderer Version", "orderer version of org Org1 invalid: version "+test.version+" is lower than required "+test.minimum)
		})
	}

	for _, version := range []string{"", "2.5.0", "3.0.0"} {
		t.Run("valid "+version, func(t *testing.T) {
			assertNoError(t, InvalidOrdererVersionFn(spec.Orderer{Version: version}, "Org1", "V2_0", "V2_5"))
		})
	}
}
