package validate

import (
	"testing"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func TestUnsupportedChaincodeLanguageFn(t *testing.T) {
	for _, language := range []string{"", spec.LanguageGo, spec.LanguageJava, spec.LanguageNode} {
		assertNoError(t, UnsupportedChaincodeLanguageFn(spec.Chaincode{Language: spec.Language{Name: language}}))
	}
	assertValidationError(t, UnsupportedChaincodeLanguageFn(spec.Chaincode{Name: "Asset", Language: spec.Language{Name: "python"}}), RuleChaincodeLanguageUnsupported, "Unsupported Chaincode Language", "chaincode Asset has unsupported language \"python\" (supported: golang, java, node)")
}

func TestEmptyChaincodeNameFn(t *testing.T) {
	assertValidationError(t, EmptyChaincodeNameFn(spec.Chaincode{}, 4), RuleChaincodeNameRequired, "Empty Chaincode Name", "name of the chaincode 4 is empty")
	assertNoError(t, EmptyChaincodeNameFn(spec.Chaincode{Name: "asset"}, 0))
}

func TestEmptyChaincodePathFn(t *testing.T) {
	assertValidationError(t, EmptyChaincodePathFn(spec.Chaincode{}, 4), RuleChaincodePathRequired, "Empty Chaincode Path", "path of the chaincode 4 is empty")
	assertNoError(t, EmptyChaincodePathFn(spec.Chaincode{Path: "chaincode/asset"}, 0))
}

func TestEmptyChaincodeVersionFn(t *testing.T) {
	assertValidationError(t, EmptyChaincodeVersionFn(spec.Chaincode{}, 4), RuleChaincodeVersionRequired, "Empty Chaincode Version", "version of the chaincode 4 is empty")
	assertNoError(t, EmptyChaincodeVersionFn(spec.Chaincode{Version: "1.0"}, 0))
}
