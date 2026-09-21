package validate

import (
	"fmt"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func DuplicatePeerNameFn(p spec.Peer, org string, seen map[string]struct{}) error {
	return duplicateValue(p.Name, seen, RulePeerNameDuplicate, "Duplicate Peer Name", fmt.Sprintf("duplicate peer name in organization %s: %s", org, p.Name))
}

func EmptyPeerNameFn(peer spec.Peer, index int, organizationName string) error {
	if peer.Name == "" {
		return &ValidationError{
			RuleID: RulePeerNameRequired,
			Rule:   "Empty Peer Name",
			Detail: fmt.Sprintf("name of the peer index %d of the organization %s is undefined", index, organizationName),
		}
	}

	return nil
}

func DuplicatePeerSubdomainFn(p spec.Peer, org string, seen map[string]struct{}) error {
	return duplicateValue(p.Subdomain, seen, RulePeerSubdomainDuplicate, "Duplicate Peer Subdomain", fmt.Sprintf("duplicate peer subdomain in organization %s: %s", org, p.Subdomain))
}

func EmptyPeerSubdomainFn(peer spec.Peer, organizationName string) error {
	if peer.Subdomain == "" {
		return &ValidationError{
			RuleID: RulePeerSubdomainRequired,
			Rule:   "Empty Peer Subdomain",
			Detail: fmt.Sprintf("subdomain of the peer %s of the organization %s is undefined", peer.Name, organizationName),
		}
	}

	return nil
}

func InvalidPeerVersionFn(peer spec.Peer, organizationName, channelCapability, applicationCapability string) error {
	if err := validateBinary(peer.Version, minimumBinaryVersion(channelCapability, applicationCapability)); err != nil {
		return &ValidationError{
			RuleID: RulePeerVersionInvalid,
			Rule:   "Invalid Peer Version",
			Detail: fmt.Sprintf("peer version of org %s invalid: %v", organizationName, err),
		}
	}

	return nil
}
