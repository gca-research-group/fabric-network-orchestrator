package generator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/yaml"
)

type MutationOperator struct {
	RuleID validate.RuleID
	Index  int
	Apply  func(node *yaml.Node)
}

type ScenarioMutation struct {
	Rule          validate.RuleID `json:"rule"`
	OperatorIndex int             `json:"operatorIndex"`
}

func (mutation *ScenarioMutation) UnmarshalJSON(data []byte) error {
	var value struct {
		Rule          validate.RuleID `json:"rule"`
		OperatorIndex *int            `json:"operatorIndex"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value.OperatorIndex == nil {
		return fmt.Errorf("operatorIndex is required")
	}
	mutation.Rule = value.Rule
	mutation.OperatorIndex = *value.OperatorIndex
	return nil
}

type ScenarioRules struct {
	Scenario  string             `json:"scenario"`
	Mutations []ScenarioMutation `json:"mutations"`
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

	manifest, err := os.Create(filepath.Join(outputDirectory, "scenarios.json"))

	if err != nil {
		return Summary{}, fmt.Errorf("create scenario manifest: %w", err)
	}

	writer := bufio.NewWriter(manifest)

	if _, err := writer.WriteString("[\n"); err != nil {
		manifest.Close()
		return Summary{}, fmt.Errorf("write scenario manifest: %w", err)
	}

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

		entry, err := json.Marshal(ScenarioRules{Scenario: scenarioName, Mutations: mutations})

		if err != nil {
			return fmt.Errorf("encode scenario %s: %w", scenarioName, err)
		}

		if summary.Total > 0 {
			if _, err := writer.WriteString(",\n"); err != nil {
				return fmt.Errorf("write scenario manifest: %w", err)
			}
		}

		if _, err := writer.WriteString("  " + string(entry)); err != nil {
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
		manifest.Close()
		return Summary{}, err
	}

	if _, err := writer.WriteString("\n]\n"); err != nil {
		manifest.Close()
		return Summary{}, fmt.Errorf("write scenario manifest: %w", err)
	}

	if err := writer.Flush(); err != nil {
		manifest.Close()
		return Summary{}, fmt.Errorf("flush scenario manifest: %w", err)
	}

	if err := manifest.Close(); err != nil {
		return Summary{}, fmt.Errorf("close scenario manifest: %w", err)
	}

	return summary, nil
}
