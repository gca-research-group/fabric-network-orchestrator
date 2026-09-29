package config

import "github.com/gca-research-group/fabric-network-orchestrator/internal/spec"

func setUpDefaultValues(config *Config) {
	for i := range config.Channels {
		for j := range config.Channels[i].Chaincodes {
			chaincode := &config.Channels[i].Chaincodes[j]
			if chaincode.Language.Name == "" {
				chaincode.Language.Name = spec.LanguageGo
			}
		}
	}

	for i := range config.Profiles {
		profile := &config.Profiles[i]
		if profile.Consensus.Type == "" {
			profile.Consensus.Type = "etcdraft"
		}
	}

	channelCapability, _ := CapabilityMap[config.Capabilities.Channel]
	applicationCapability, _ := CapabilityMap[config.Capabilities.Application]
	ordererCapability, _ := CapabilityMap[config.Capabilities.Orderer]

	hasBootstrapOrganization := false

	for i := range config.Organizations {
		organization := &config.Organizations[i]

		if organization.CertificateAuthority.Version == "" {
			organization.CertificateAuthority.Version = "latest"
		}

		for j := range organization.Orderers {

			orderer := &organization.Orderers[j]

			if orderer.Port == 0 {
				orderer.Port = 7050
			}

			if orderer.Version == "" {
				if channelCapability > ordererCapability {
					orderer.Version = DefaultVersionByCapability[config.Capabilities.Channel]
				} else {
					orderer.Version = DefaultVersionByCapability[config.Capabilities.Orderer]
				}
			}
		}

		if organization.Bootstrap {
			hasBootstrapOrganization = true
		}

		hasAnchorPeer := false

		for i := range organization.Peers {
			peer := &organization.Peers[i]

			if peer.IsAnchor {
				hasAnchorPeer = true
			}

			if peer.Version == "" {
				if channelCapability > applicationCapability {
					peer.Version = DefaultVersionByCapability[config.Capabilities.Channel]
				} else {
					peer.Version = DefaultVersionByCapability[config.Capabilities.Application]
				}
			}
		}

		if !hasAnchorPeer && len(organization.Peers) > 0 {
			organization.Peers[0].IsAnchor = true
		}
	}

	if !hasBootstrapOrganization {
		for i, organization := range config.Organizations {
			if len(organization.Orderers) != 0 {
				config.Organizations[i].Bootstrap = true
				break
			}
		}
	}
}
