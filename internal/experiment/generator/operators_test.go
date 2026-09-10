package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/format"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/experiment/seed"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/yaml"
	yamlv3 "gopkg.in/yaml.v3"
)

const testSeedYAML = `
output: output/example
network: example

capabilities:
  channel: V2_0
  orderer: V2_0
  application: V2_5

organizations:
  - name: Org1
    bootstrap: true
    domain: org1.example.com
    orderers:
      - name: Orderer
        subdomain: orderer
    peers:
      - name: Peer0
        subdomain: peer0
        exposePort: 7051
    certificateAuthority:
      exposePort: 7054

  - name: Org2
    bootstrap: false
    domain: org2.example.com
    peers:
      - name: Peer0
        subdomain: peer0
        exposePort: 8051

  - name: Org3
    domain: org3.example.com
    peers:
      - name: Peer0
        subdomain: peer0
        exposePort: 9051

profiles:
  - name: DefaultProfile
    organizations:
      - Org1
      - Org2
      - Org3

channels:
  - name: defaultchannel
    profile: DefaultProfile
    chaincodes:
      - name: Asset
        version: "1.0"
        path: samples/chaincodes/asset
        language:
          name: golang
          version: "1.26"

      - name: Product
        version: "1.0"
        path: samples/chaincodes/product

      - name: PrivateAgreement
        version: "1.0"
        path: samples/chaincodes/private-agreement
`

var allRuleIDs = []validate.RuleID{
	validate.RuleOutputDirectoryNameInvalid,
	validate.RuleNetworkNameRequired,
	validate.RuleNetworkNameInvalid,
	validate.RuleOrganizationsRequired,
	validate.RuleOrganizationNameRequired,
	validate.RuleOrganizationDomainRequired,
	validate.RuleOrganizationDomainDuplicate,
	validate.RuleOrganizationUsersInvalid,
	validate.RuleDomainInvalid,
	validate.RuleCertificateAuthorityPortInvalid,
	validate.RulePeerNameRequired,
	validate.RulePeerSubdomainRequired,
	validate.RulePeerPortInvalid,
	validate.RulePeerInternalPortInvalid,
	validate.RulePeerNameDuplicate,
	validate.RulePeerSubdomainDuplicate,
	validate.RuleOrdererNameRequired,
	validate.RuleOrdererSubdomainRequired,
	validate.RuleOrdererPortInvalid,
	validate.RuleOrdererInternalPortInvalid,
	validate.RuleOrdererNameDuplicate,
	validate.RuleChaincodeNameRequired,
	validate.RuleChaincodePathRequired,
	validate.RuleChaincodeVersionRequired,
	validate.RuleChaincodeNameDuplicate,
	validate.RuleProfileOrganizationsRequired,
	validate.RuleProfileNameRequired,
	validate.RuleProfileNameDuplicate,
	validate.RuleChannelProfileRequired,
	validate.RuleChannelNameRequired,
	validate.RuleChannelNameInvalid,
	validate.RuleChannelNameDuplicate,
	validate.RuleChannelProfileUndefined,
	validate.RuleChannelCapabilityUnsupported,
	validate.RuleApplicationCapabilityUnsupported,
	validate.RuleOrdererCapabilityUnsupported,
	validate.RuleOrganizationNameDuplicate,
	validate.RulePeerVersionInvalid,
	validate.RuleOrdererVersionInvalid,
	validate.RuleOrdererTopologyRequired,
	validate.RuleBootstrapOrganizationsMultiple,
	validate.RuleConsensusTypeInvalid,
	validate.RuleProfileOrganizationUndefined,
	validate.RuleExposedPortConflict,
}

func seedNode(t *testing.T) *yaml.Node {
	t.Helper()
	node, err := yaml.FromBytes([]byte(testSeedYAML))
	if err != nil {
		t.Fatalf("parse seed: %v", err)
	}
	return node
}

func decodeConfig(t *testing.T, node *yaml.Node) spec.Config {
	t.Helper()
	var configuration spec.Config
	if err := (*yamlv3.Node)(node.Document()).Decode(&configuration); err != nil {
		t.Fatalf("decode mutation: %v", err)
	}
	return configuration
}

func TestSeedIsValid(t *testing.T) {
	configuration := decodeConfig(t, seedNode(t))
	if err := validate.Config(configuration); err != nil {
		t.Fatalf("seed must be valid (decoded output %q): %v", configuration.Output, err)
	}
}

func TestBundledSeedSupportsEveryOperator(t *testing.T) {
	for _, group := range operators {
		for index, operator := range group {
			t.Run(fmt.Sprintf("%s/%d", operator.RuleID, index), func(t *testing.T) {
				node, err := yaml.FromBytes([]byte(seed.YAML))
				if err != nil {
					t.Fatal(err)
				}
				operator.Apply(node.Document())
				for _, validationError := range validate.Errors(validate.Config(decodeConfig(t, node))) {
					if validationError.RuleID == operator.RuleID {
						return
					}
				}
				t.Fatalf("operator did not trigger %s", operator.RuleID)
			})
		}
	}
}

func TestEveryOperatorTriggersDeclaredRule(t *testing.T) {
	for _, group := range operators {
		for index, operator := range group {
			operator := operator
			t.Run(string(operator.RuleID)+"/"+string(rune('0'+index)), func(t *testing.T) {
				node := seedNode(t)
				operator.Apply(node.Document())
				err := validate.Config(decodeConfig(t, node))
				var validationError *validate.ValidationError
				if !errors.As(err, &validationError) {
					t.Fatalf("expected ValidationError, got %v", err)
				}
				if validationError.RuleID != operator.RuleID {
					t.Fatalf("operator declares %s, triggered %s", operator.RuleID, validationError.RuleID)
				}
			})
		}
	}
}

func TestOperatorRegistryCoversEveryRuleOnce(t *testing.T) {
	registered := make(map[validate.RuleID]int)
	for _, group := range operators {
		if len(group) == 0 {
			t.Fatal("operator registry contains an empty group")
		}
		groupRule := group[0].RuleID
		for _, operator := range group {
			if operator.RuleID != groupRule {
				t.Fatalf("group %s also contains %s", groupRule, operator.RuleID)
			}
		}
		registered[groupRule]++
	}
	for _, ruleID := range allRuleIDs {
		if registered[ruleID] != 1 {
			t.Errorf("rule %s registered %d times", ruleID, registered[ruleID])
		}
		delete(registered, ruleID)
	}
	for ruleID := range registered {
		t.Errorf("unexpected registered rule %s", ruleID)
	}
}

func TestGenerateCombinationsUsesAtMostOneOperatorPerRule(t *testing.T) {
	err := WalkCombinations(operators, 3, incompatibilities, func(combination []MutationOperator) error {
		seen := make(map[validate.RuleID]bool)
		for _, operator := range combination {
			if seen[operator.RuleID] {
				return fmt.Errorf("combination contains rule %s more than once", operator.RuleID)
			}
			seen[operator.RuleID] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWalkCombinationsStreamsOneHundredThousandCombinations(t *testing.T) {
	groups := make([][]MutationOperator, 5)
	for groupIndex := range groups {
		for operatorIndex := 0; operatorIndex < 10; operatorIndex++ {
			groups[groupIndex] = append(groups[groupIndex], MutationOperator{RuleID: validate.RuleID(fmt.Sprintf("rule-%d-%d", groupIndex, operatorIndex))})
		}
	}

	count := 0
	if err := WalkCombinations(groups, 5, nil, func([]MutationOperator) error {
		count++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 100_000 {
		t.Fatalf("visited %d combinations, expected 100000", count)
	}
}

func TestCountCombinationsMatchesWalk(t *testing.T) {
	groups := [][]MutationOperator{
		{{RuleID: "one-a"}, {RuleID: "one-b"}},
		{{RuleID: "two-a"}},
		{{RuleID: "three-a"}, {RuleID: "three-b"}},
	}

	want := 0
	if err := WalkCombinations(groups, 2, nil, func([]MutationOperator) error {
		want++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	got, err := countCombinations(groups, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count = %d, expected %d", got, want)
	}
}

func TestGenerateStreamsValidManifest(t *testing.T) {
	outputDirectory := t.TempDir()
	seedYAML := strings.Replace(testSeedYAML, "network: example", "network: custom-seed", 1)
	var progress []int
	progressTotal := -1
	summary, err := Generate([]byte(seedYAML), 1, outputDirectory, func(completed, total int) {
		if progressTotal != -1 && total != progressTotal {
			t.Fatalf("progress total changed from %d to %d", progressTotal, total)
		}
		progressTotal = total
		progress = append(progress, completed)
	})
	if err != nil {
		t.Fatal(err)
	}

	scenarios, err := parquet.ReadFile[ScenarioRules](filepath.Join(outputDirectory, "scenarios.parquet"))
	if err != nil {
		t.Fatalf("decode generated manifest: %v", err)
	}
	if summary.Total != len(scenarios) || summary.Total == 0 {
		t.Fatalf("summary total %d does not match manifest length %d", summary.Total, len(scenarios))
	}
	assertZstdParquet(t, filepath.Join(outputDirectory, "scenarios.parquet"))
	if progressTotal != summary.Total || len(progress) != summary.Total+1 {
		t.Fatalf("progress reported %d updates with total %d for %d scenarios", len(progress), progressTotal, summary.Total)
	}
	for completed, reported := range progress {
		if reported != completed {
			t.Fatalf("progress update %d reported %d", completed, reported)
		}
	}
	for _, scenario := range scenarios {
		if len(scenario.Mutations) != 1 {
			t.Fatalf("scenario %s has %d mutations", scenario.Scenario, len(scenario.Mutations))
		}
		operator, found := FindMutationOperator(scenario.Mutations[0])
		if !found || operator.RuleID != scenario.Mutations[0].Rule || operator.Index != scenario.Mutations[0].OperatorIndex {
			t.Fatalf("scenario %s has an invalid operator reference: %+v", scenario.Scenario, scenario.Mutations[0])
		}
	}
	if _, err := os.Stat(filepath.Join(outputDirectory, "config")); !os.IsNotExist(err) {
		t.Fatalf("generation created a config directory: %v", err)
	}
	file, err := os.Open(filepath.Join(outputDirectory, "scenarios.parquet"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		t.Fatal(err)
	}
	document, err := parquet.OpenFile(file, info.Size())
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if embedded, found := document.Lookup(SeedYAMLMetadataKey); !found || embedded != seedYAML {
		t.Fatal("scenario manifest does not contain the exact supplied seed YAML")
	}
	if summary.MutationOperatorsUsed != summary.Total {
		t.Fatalf("used %d mutation operators for %d one-mutation scenarios", summary.MutationOperatorsUsed, summary.Total)
	}
}

func TestReplayScenarioAppliesRecordedMutations(t *testing.T) {
	seed := seedNode(t)
	mutation := ScenarioMutation{Rule: operators[0][0].RuleID, OperatorIndex: operators[0][0].Index}
	replayed, err := ReplayScenario(seed, ScenarioRules{Scenario: "000001", Mutations: []ScenarioMutation{mutation}})
	if err != nil {
		t.Fatal(err)
	}
	direct := seed.Clone()
	operators[0][0].Apply(direct.Document())
	replayedYAML, _ := replayed.ToBytes()
	directYAML, _ := direct.ToBytes()
	if string(replayedYAML) != string(directYAML) {
		t.Fatal("replayed mutation differs from direct application")
	}
}

func assertZstdParquet(t *testing.T, path string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	document, err := parquet.OpenFile(file, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range document.Metadata().RowGroups {
		for _, column := range group.Columns {
			if column.MetaData.Codec != format.Zstd {
				t.Fatalf("column %v uses %s compression", column.MetaData.PathInSchema, column.MetaData.Codec)
			}
		}
	}
}

func TestMutationReferenceReproducesMutation(t *testing.T) {
	for _, group := range operators {
		for _, want := range group {
			reference := ScenarioMutation{Rule: want.RuleID, OperatorIndex: want.Index}
			got, found := FindMutationOperator(reference)
			if !found {
				t.Fatalf("operator reference was not resolved: %+v", reference)
			}
			wantNode, gotNode := seedNode(t), seedNode(t)
			want.Apply(wantNode.Document())
			got.Apply(gotNode.Document())
			if !reflect.DeepEqual(decodeConfig(t, gotNode), decodeConfig(t, wantNode)) {
				t.Fatalf("operator reference %+v produced a different mutation", reference)
			}
		}
	}
}

func TestGenerateFromDocumentedSample(t *testing.T) {
	seedYAML, err := os.ReadFile(filepath.Join("..", "..", "..", "samples", "network-with-chaincode.yml"))
	if err != nil {
		t.Fatal(err)
	}

	summary, err := Generate(seedYAML, 1, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total == 0 {
		t.Fatal("expected scenarios from documented sample seed")
	}
}

func TestGenerateCountsEveryAppliedMutationOperator(t *testing.T) {
	originalOperators := operators
	operators = indexOperators([][]MutationOperator{{networkNameInvalidOperators[0]}, {organizationsRequiredOperators[0]}})
	defer func() { operators = originalOperators }()

	summary, err := Generate([]byte(testSeedYAML), 2, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 1 || summary.MutationOperatorsUsed != 2 {
		t.Fatalf("unexpected generation summary: %+v", summary)
	}
}

func TestIncompatibilitiesAreSymmetric(t *testing.T) {
	for first, incompatibleRules := range incompatibilities {
		for _, second := range incompatibleRules {
			if !rulesConflict(first, second, incompatibilities) || !rulesConflict(second, first, incompatibilities) {
				t.Fatalf("rules %s and %s must conflict in both directions", first, second)
			}

			firstOperator := MutationOperator{RuleID: first}
			secondOperator := MutationOperator{RuleID: second}
			for _, groups := range [][][]MutationOperator{
				{{firstOperator}, {secondOperator}},
				{{secondOperator}, {firstOperator}},
			} {
				count := 0
				if err := WalkCombinations(groups, 2, incompatibilities, func([]MutationOperator) error { count++; return nil }); err != nil {
					t.Fatal(err)
				}
				if count != 0 {
					t.Fatalf("generated rules %s and %s in an incompatible order", first, second)
				}
			}
		}
	}
}

func TestGeneratedCombinationsExcludeIncompatibleRules(t *testing.T) {
	represented := make(map[validate.RuleID]bool)
	err := WalkCombinations(operators, DefaultMutationCount, incompatibilities, func(combination []MutationOperator) error {
		for i, first := range combination {
			represented[first.RuleID] = true
			for _, second := range combination[i+1:] {
				if rulesConflict(first.RuleID, second.RuleID, incompatibilities) {
					return fmt.Errorf("generated incompatible rules %s and %s", first.RuleID, second.RuleID)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, ruleID := range allRuleIDs {
		if !represented[ruleID] {
			t.Errorf("rule %s is not represented by any generated combination", ruleID)
		}
	}
}

func TestGeneratedCombinationsTriggerEveryDeclaredRule(t *testing.T) {
	index := 0
	err := WalkCombinations(operators, DefaultMutationCount, incompatibilities, func(combination []MutationOperator) error {
		node := seedNode(t)
		for _, operator := range combination {
			operator.Apply(node.Document())
		}

		actual := make(map[validate.RuleID]bool)
		for _, validationError := range validate.Errors(validate.Config(decodeConfig(t, node))) {
			actual[validationError.RuleID] = true
		}
		for _, operator := range combination {
			if !actual[operator.RuleID] {
				return fmt.Errorf("combination %d (%v) did not trigger declared rule %s", index, combinationRuleIDs(combination), operator.RuleID)
			}
		}
		index++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func combinationRuleIDs(combination []MutationOperator) []validate.RuleID {
	rules := make([]validate.RuleID, 0, len(combination))
	for _, operator := range combination {
		rules = append(rules, operator.RuleID)
	}
	return rules
}

func TestOrganizationsRequiredExcludesOrganizationDomainRequired(t *testing.T) {
	err := WalkCombinations(operators, DefaultMutationCount, incompatibilities, func(combination []MutationOperator) error {
		hasOrganizationsRequired := false
		hasOrganizationDomainRequired := false
		for _, operator := range combination {
			hasOrganizationsRequired = hasOrganizationsRequired || operator.RuleID == validate.RuleOrganizationsRequired
			hasOrganizationDomainRequired = hasOrganizationDomainRequired || operator.RuleID == validate.RuleOrganizationDomainRequired
		}
		if hasOrganizationsRequired && hasOrganizationDomainRequired {
			return errors.New("generated organizations.required with organization.domain.required")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
