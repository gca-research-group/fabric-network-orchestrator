package validate

import (
	"fmt"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func DuplicateOrdererNameFn(o spec.Orderer, org string, seen map[string]struct{}) error {
	return duplicateValue(o.Name, seen, RuleOrdererNameDuplicate, "Duplicate Orderer Name", fmt.Sprintf("duplicate orderer name in organization %s: %s", org, o.Name))
}

func EmptyOrdererNameFn(orderer spec.Orderer, index int, organizationName string) error {
	if orderer.Name == "" {
		return &ValidationError{
			RuleID: RuleOrdererNameRequired,
			Rule:   "Empty Orderer Name",
			Detail: fmt.Sprintf("name of the orderer index %d of the organization %s is undefined", index, organizationName),
		}
	}

	return nil
}

func EmptyOrdererSubdomainFn(orderer spec.Orderer, organizationName string) error {
	if orderer.Subdomain == "" {
		return &ValidationError{
			RuleID: RuleOrdererSubdomainRequired,
			Rule:   "Empty Orderer Subdomain",
			Detail: fmt.Sprintf("subdomain of the orderer %s of the organization %s is undefined", orderer.Name, organizationName),
		}
	}

	return nil
}

func InvalidOrdererVersionFn(orderer spec.Orderer, organizationName, channelCapability, ordererCapability string) error {
	if err := validateBinary(orderer.Version, minimumBinaryVersion(channelCapability, ordererCapability)); err != nil {
		return &ValidationError{
			RuleID: RuleOrdererVersionInvalid,
			Rule:   "Invalid Orderer Version",
			Detail: fmt.Sprintf("orderer version of org %s invalid: %v", organizationName, err),
		}
	}

	return nil
}
