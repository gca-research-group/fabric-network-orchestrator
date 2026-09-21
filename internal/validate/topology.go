package validate

import (
	"errors"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func NoOrdererTopologyFn(hasOrderer bool) error {
	if !hasOrderer {
		return &ValidationError{
			RuleID: RuleOrdererTopologyRequired,
			Rule:   "No Orderer Topology",
			Detail: "at least one orderer must be configured",
		}
	}

	return nil
}

func ValidateTopologyFn(configuration spec.Config) error {
	var validationErrors []error
	collect := func(err error) {
		if err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	organizationNames := make(map[string]struct{})
	organizationDomains := make(map[string]struct{})
	exposedPorts := make(map[int]portOwner)
	bootstrapCount := 0
	hasOrderer := false

	for _, organization := range configuration.Organizations {
		collect(DuplicateOrganizationNameFn(organization, organizationNames))
		collect(DuplicateOrganizationDomainFn(organization, organizationDomains))

		organizationNames[organization.Name] = struct{}{}

		if organization.Bootstrap {
			bootstrapCount++
		}

		if len(organization.Orderers) > 0 {
			hasOrderer = true
		}

		collect(ExposedPortConflictFn(organization.CertificateAuthority.ExposePort, portOwner{ownerType: "Certificate Authority", name: organization.Name}, exposedPorts))

		peerNames := make(map[string]struct{})
		peerSubdomains := make(map[string]struct{})
		for _, peer := range organization.Peers {
			collect(DuplicatePeerNameFn(peer, organization.Name, peerNames))
			collect(DuplicatePeerSubdomainFn(peer, organization.Name, peerSubdomains))
			collect(InvalidPeerVersionFn(peer, organization.Name, configuration.Capabilities.Channel, configuration.Capabilities.Application))
			collect(ExposedPortConflictFn(peer.ExposePort, portOwner{ownerType: "peer", name: peer.Name}, exposedPorts))
		}

		ordererNames := make(map[string]struct{})
		for _, orderer := range organization.Orderers {
			collect(DuplicateOrdererNameFn(orderer, organization.Name, ordererNames))
			collect(InvalidOrdererVersionFn(orderer, organization.Name, configuration.Capabilities.Channel, configuration.Capabilities.Orderer))
			collect(ExposedPortConflictFn(orderer.ExposePort, portOwner{ownerType: "orderer", name: orderer.Name}, exposedPorts))
		}
	}

	collect(NoOrdererTopologyFn(hasOrderer))
	collect(MultipleBootstrapOrganizationsFn(bootstrapCount))

	for _, profile := range configuration.Profiles {
		collect(InvalidConsensusTypeFn(profile))
		collect(UndefinedProfileOrganizationFn(profile, organizationNames))
	}

	return errors.Join(validationErrors...)
}
