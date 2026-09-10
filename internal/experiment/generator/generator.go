package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
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

const SeedYAMLMetadataKey = "fabric-network-orchestrator.seed-yaml"

func Generate(seedYAML []byte, mutationCount int, outputDirectory string, progress ProgressFunc) (Summary, error) {
	if _, err := yaml.FromBytes(seedYAML); err != nil {
		return Summary{}, fmt.Errorf("parse seed configuration: %w", err)
	}
	if _, err := config.LoadConfigFromYAML(seedYAML); err != nil {
		return Summary{}, fmt.Errorf("validate seed configuration: %w", err)
	}

	total, err := countCombinations(operators, mutationCount, incompatibilities)
	if err != nil {
		return Summary{}, fmt.Errorf("count scenarios: %w", err)
	}
	if progress != nil {
		progress(0, total)
	}

	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return Summary{}, fmt.Errorf("create output directory: %w", err)
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
	writer.SetKeyValueMetadata(SeedYAMLMetadataKey, string(seedYAML))

	summary := Summary{}

	err = WalkCombinations(operators, mutationCount, incompatibilities, func(scenario []MutationOperator) error {
		mutations := make([]ScenarioMutation, 0, len(scenario))

		for _, operator := range scenario {
			mutations = append(mutations, ScenarioMutation{Rule: operator.RuleID, OperatorIndex: operator.Index})
		}

		scenarioName := fmt.Sprintf("%06d", summary.Total+1)

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

// ReplayScenario clones the seed and applies the scenario's recorded mutations in order.
func ReplayScenario(seed *yaml.Node, scenario ScenarioRules) (*yaml.Node, error) {
	clone := seed.Clone()
	document := clone.Document()
	for _, mutation := range scenario.Mutations {
		operator, found := FindMutationOperator(mutation)
		if !found {
			return nil, fmt.Errorf("unknown mutation operator %s/%d", mutation.Rule, mutation.OperatorIndex)
		}
		operator.Apply(document)
	}
	return clone, nil
}
