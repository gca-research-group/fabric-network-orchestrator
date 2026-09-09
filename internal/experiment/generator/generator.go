package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/yaml"
)

type MutationOperator struct {
	RuleID validate.RuleID
	Index  int
	Apply  func(node *yaml.Node)
}

type ScenarioMutation struct {
	Rule          validate.RuleID `parquet:"rule,dict"`
	OperatorIndex int             `parquet:"operatorIndex"`
}

type ScenarioRules struct {
	Scenario  string             `parquet:"scenario,dict"`
	Mutations []ScenarioMutation `parquet:"mutations,list"`
}

type Summary struct {
	Total                 int
	MutationOperatorsUsed int
}

type ProgressFunc func(completed, total int)

var operators = indexOperators([][]MutationOperator{
	outputDirectoryOperators,
	networkNameInvalidOperators,
	organizationDomainDuplicateOperators,
	domainInvalidOperators,
	organizationUsersInvalidOperators,
	peerNameDuplicateOperators,
	peerSubdomainDuplicateOperators,
	peerInternalPortInvalidOperators,
	ordererNameDuplicateOperators,
	ordererInternalPortInvalidOperators,
	profileNameDuplicateOperators,
	channelProfileUndefinedOperators,
	channelNameDuplicateOperators,
	chaincodeNameDuplicateOperators,
	organizationDomainRequiredOperators,
	certificateAuthorityPortInvalidOperators,
	peerSubdomainRequiredOperators,
	peerPortInvalidOperators,
	ordererSubdomainRequiredOperators,
	ordererPortInvalidOperators,
	chaincodePathRequiredOperators,
	chaincodeVersionRequiredOperators,
	channelProfileRequiredOperators,
	channelNameInvalidOperators,
	channelCapabilityUnsupportedOperators,
	applicationCapabilityUnsupportedOperators,
	ordererCapabilityUnsupportedOperators,
	peerVersionInvalidOperators,
	ordererVersionInvalidOperators,
	bootstrapOrganizationsMultipleOperators,
	consensusTypeInvalidOperators,
	profileOrganizationUndefinedOperators,
	exposedPortConflictOperators,
	organizationNameDuplicateOperators,
	peerNameRequiredOperators,
	ordererNameRequiredOperators,
	chaincodeNameRequiredOperators,
	channelNameRequiredOperators,
	ordererTopologyRequiredOperators,
	organizationNameRequiredOperators,
	profileOrganizationsRequiredOperators,
	profileNameRequiredOperators,
	networkNameRequiredOperators,
	organizationsRequiredOperators,
})

func indexOperators(groups [][]MutationOperator) [][]MutationOperator {
	for groupIndex := range groups {
		for operatorIndex := range groups[groupIndex] {
			groups[groupIndex][operatorIndex].Index = operatorIndex
		}
	}
	return groups
}

func FindMutationOperator(mutation ScenarioMutation) (MutationOperator, bool) {
	for _, group := range operators {
		if len(group) > 0 && group[0].RuleID == mutation.Rule && mutation.OperatorIndex >= 0 && mutation.OperatorIndex < len(group) {
			return group[mutation.OperatorIndex], true
		}
	}
	return MutationOperator{}, false
}

var incompatibilities = IncompatibilityPolicy{
	validate.RuleOrganizationsRequired: {
		validate.RuleOrganizationDomainDuplicate,
		validate.RuleDomainInvalid,
		validate.RuleOrganizationUsersInvalid,
		validate.RuleOrganizationNameRequired,
		validate.RuleOrganizationDomainRequired,
		validate.RuleCertificateAuthorityPortInvalid,
		validate.RulePeerNameRequired,
		validate.RulePeerSubdomainRequired,
		validate.RulePeerPortInvalid,
		validate.RuleOrdererNameRequired,
		validate.RuleOrdererSubdomainRequired,
		validate.RuleOrdererPortInvalid,
		validate.RuleOrganizationNameDuplicate,
		validate.RulePeerVersionInvalid,
		validate.RuleOrdererVersionInvalid,
		validate.RuleBootstrapOrganizationsMultiple,
		validate.RuleExposedPortConflict,
		validate.RulePeerNameDuplicate,
		validate.RulePeerSubdomainDuplicate,
		validate.RulePeerInternalPortInvalid,
		validate.RuleOrdererNameDuplicate,
		validate.RuleOrdererInternalPortInvalid,
	},
	validate.RuleOrdererTopologyRequired: {
		validate.RuleOrdererNameRequired,
		validate.RuleOrdererNameDuplicate,
		validate.RuleOrdererSubdomainRequired,
		validate.RuleOrdererPortInvalid,
		validate.RuleOrdererInternalPortInvalid,
		validate.RuleOrdererVersionInvalid,
	},
	validate.RuleChannelCapabilityUnsupported: {
		validate.RulePeerVersionInvalid,
		validate.RuleOrdererVersionInvalid,
	},
	validate.RuleProfileOrganizationsRequired: {
		validate.RuleProfileOrganizationUndefined,
	},
	validate.RulePeerPortInvalid: {
		validate.RuleExposedPortConflict,
	},
	validate.RuleOrganizationNameRequired: {
		validate.RuleOrganizationNameDuplicate,
	},
	validate.RuleProfileNameRequired: {
		validate.RuleProfileNameDuplicate,
		validate.RuleChannelProfileUndefined,
	},
	validate.RuleChannelNameRequired: {
		validate.RuleChannelNameDuplicate,
		validate.RuleChannelNameInvalid,
	},
	validate.RuleChannelProfileUndefined: {
		validate.RuleChannelProfileRequired,
	},
	validate.RuleNetworkNameRequired: {
		validate.RuleNetworkNameInvalid,
	},
	validate.RuleOrganizationDomainRequired: {
		validate.RuleDomainInvalid,
		validate.RuleOrganizationDomainDuplicate,
	},
	validate.RuleOrganizationDomainDuplicate: {
		validate.RuleDomainInvalid,
	},
	validate.RulePeerNameDuplicate: {
		validate.RulePeerNameRequired,
	},
	validate.RulePeerSubdomainDuplicate: {
		validate.RulePeerSubdomainRequired,
	},
	validate.RuleOrdererNameDuplicate: {
		validate.RuleOrdererNameRequired,
	},
	validate.RuleChaincodeNameDuplicate: {
		validate.RuleChaincodeNameRequired,
	},
}

const DefaultMutationCount = 3

func Generate(seedYAML []byte, mutationCount int, outputDirectory string, progress ProgressFunc) (Summary, error) {
	seed, err := yaml.FromBytes(seedYAML)
	if err != nil {
		return Summary{}, fmt.Errorf("parse seed configuration: %w", err)
	}

	total, err := countCombinations(operators, mutationCount, incompatibilities)
	if err != nil {
		return Summary{}, fmt.Errorf("count scenarios: %w", err)
	}
	if progress != nil {
		progress(0, total)
	}

	configDirectory := filepath.Join(outputDirectory, "config")

	if err := os.MkdirAll(configDirectory, 0755); err != nil {
		return Summary{}, fmt.Errorf("create config directory: %w", err)
	}

	manifest, err := os.CreateTemp(outputDirectory, ".scenarios-*.parquet")
	if err != nil {
		return Summary{}, fmt.Errorf("create scenario manifest: %w", err)
	}
	defer os.Remove(manifest.Name())
	writer := parquet.NewGenericWriter[ScenarioRules](manifest,
		parquet.Compression(&zstd.Codec{}),
		parquet.MaxRowsPerRowGroup(64*1024),
	)

	summary := Summary{}

	err = WalkCombinations(operators, mutationCount, incompatibilities, func(scenario []MutationOperator) error {
		clone := seed.Clone()
		doc := clone.Document()
		mutations := make([]ScenarioMutation, 0, len(scenario))

		for _, operator := range scenario {
			mutations = append(mutations, ScenarioMutation{Rule: operator.RuleID, OperatorIndex: operator.Index})
			operator.Apply(doc)
		}

		scenarioName := fmt.Sprintf("%06d", summary.Total+1)

		if err := clone.ToFile(filepath.Join(configDirectory, scenarioName+".yaml")); err != nil {
			return fmt.Errorf("write scenario %s: %w", scenarioName, err)
		}

		if _, err := writer.Write([]ScenarioRules{{Scenario: scenarioName, Mutations: mutations}}); err != nil {
			return fmt.Errorf("write scenario manifest: %w", err)
		}

		summary.Total++
		summary.MutationOperatorsUsed += len(mutations)
		if progress != nil {
			progress(summary.Total, total)
		}

		return nil
	})

	if err != nil {
		return Summary{}, errors.Join(err, writer.Close(), manifest.Close())
	}
	if err := errors.Join(writer.Close(), manifest.Close()); err != nil {
		return Summary{}, fmt.Errorf("close scenario manifest: %w", err)
	}
	if err := os.Rename(manifest.Name(), filepath.Join(outputDirectory, "scenarios.parquet")); err != nil {
		return Summary{}, fmt.Errorf("replace scenario manifest: %w", err)
	}

	return summary, nil
}
