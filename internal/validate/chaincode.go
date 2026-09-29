package validate

import (
	"fmt"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func DuplicateChaincodeNameFn(c spec.Chaincode, channel string, seen map[string]struct{}) error {
	return duplicateValue(c.Name, seen, RuleChaincodeNameDuplicate, "Duplicate Chaincode Name", fmt.Sprintf("duplicate chaincode name in channel %s: %s", channel, c.Name))
}

func EmptyChaincodeNameFn(chaincode spec.Chaincode, index int) error {
	if chaincode.Name == "" {
		return &ValidationError{
			RuleID: RuleChaincodeNameRequired,
			Rule:   "Empty Chaincode Name",
			Detail: fmt.Sprintf("name of the chaincode %d is empty", index),
		}
	}

	return nil
}

func EmptyChaincodePathFn(chaincode spec.Chaincode, index int) error {
	if chaincode.Path == "" {
		return &ValidationError{RuleID: RuleChaincodePathRequired, Rule: "Empty Chaincode Path", Detail: fmt.Sprintf("path of the chaincode %d is empty", index)}
	}

	return nil
}

func EmptyChaincodeVersionFn(chaincode spec.Chaincode, index int) error {
	if chaincode.Version == "" {
		return &ValidationError{
			RuleID: RuleChaincodeVersionRequired,
			Rule:   "Empty Chaincode Version",
			Detail: fmt.Sprintf("version of the chaincode %d is empty", index),
		}
	}

	return nil
}

func UnsupportedChaincodeLanguageFn(chaincode spec.Chaincode) error {
	switch chaincode.Language.Name {
	case "", spec.LanguageGo, spec.LanguageJava, spec.LanguageNode:
		return nil
	default:
		return validationError(RuleChaincodeLanguageUnsupported, "Unsupported Chaincode Language", fmt.Sprintf("chaincode %s has unsupported language %q (supported: golang, java, node)", chaincode.Name, chaincode.Language.Name))
	}
}
