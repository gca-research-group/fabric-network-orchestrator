package validate

import (
	"errors"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func Config(configuration spec.Config) error {
	var validationErrors []error
	collect := func(err error) {
		if err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	collect(InvalidOutputDirectoryNameFn(configuration))
	collect(EmptyNetworkNameFn(configuration))
	collect(InvalidNetworkNameFn(configuration))
	collect(NoOrganizationFn(configuration))

	for i, organization := range configuration.Organizations {
		collect(EmptyOrganizationNameFn(organization, i))
		collect(EmptyOrganizationDomainFn(organization, i))
		collect(InvalidDomainFn(organization))
		collect(InvalidOrganizationUsersFn(organization))
		collect(InvalidCertificateAuthorityPortFn(organization))

		for j, peer := range organization.Peers {
			collect(EmptyPeerNameFn(peer, j, organization.Name))
			collect(EmptyPeerSubdomainFn(peer, organization.Name))
			collect(InvalidPeerPortFn(peer, organization.Name))
			collect(InvalidPeerInternalPortFn(peer, organization.Name))
		}

		for j, orderer := range organization.Orderers {
			collect(EmptyOrdererNameFn(orderer, j, organization.Name))
			collect(EmptyOrdererSubdomainFn(orderer, organization.Name))
			collect(InvalidOrdererPortFn(orderer, organization.Name))
			collect(InvalidOrdererInternalPortFn(orderer, organization.Name))
		}
	}

	profileNames := make(map[string]struct{})
	for i, profile := range configuration.Profiles {
		collect(EmptyProfileNameFn(profile, i))
		collect(DuplicateProfileNameFn(profile, profileNames))
		collect(EmptyProfileOrganizationsFn(profile))
	}

	channelNames := make(map[string]struct{})
	for _, channel := range configuration.Channels {
		collect(EmptyChannelNameFn(channel))
		collect(EmptyChannelProfileFn(channel))
		collect(InvalidChannelNameFn(channel))
		collect(DuplicateChannelNameFn(channel, channelNames))
		collect(UndefinedChannelProfileFn(channel, profileNames))

		chaincodeNames := make(map[string]struct{})
		for i, chaincode := range channel.Chaincodes {
			collect(EmptyChaincodeNameFn(chaincode, i))
			collect(EmptyChaincodePathFn(chaincode, i))
			collect(EmptyChaincodeVersionFn(chaincode, i))
			collect(UnsupportedChaincodeLanguageFn(chaincode))
			collect(DuplicateChaincodeNameFn(chaincode, channel.Name, chaincodeNames))
		}
	}

	collect(UnsupportedChannelCapabilityFn(configuration))
	collect(UnsupportedApplicationCapabilityFn(configuration))
	collect(UnsupportedOrdererCapabilityFn(configuration))
	collect(ValidateTopologyFn(configuration))

	return errors.Join(validationErrors...)
}
