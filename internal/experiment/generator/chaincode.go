package generator

import (
	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/yaml"
)

var chaincodeNameDuplicateOperators = []MutationOperator{{RuleID: validate.RuleChaincodeNameDuplicate, Apply: func(n *yaml.Node) {
	chaincode(n, "defaultchannel", "Product").GetValue("name").SetScalar("Asset", yaml.StringType)
}}}

var chaincodeNameRequiredOperators = []MutationOperator{
	{
		RuleID: validate.RuleChaincodeNameRequired,
		Apply: func(node *yaml.Node) {
			chaincode(node, "defaultchannel", "Asset").GetValue("name").SetScalar("", yaml.StringType)
		},
	},
}

var chaincodePathRequiredOperators = []MutationOperator{{
	RuleID: validate.RuleChaincodePathRequired,
	Apply: func(node *yaml.Node) {
		chaincode(node, "defaultchannel", "Asset").GetValue("path").SetScalar("", yaml.StringType)
	},
}}

var chaincodeVersionRequiredOperators = []MutationOperator{{
	RuleID: validate.RuleChaincodeVersionRequired,
	Apply: func(node *yaml.Node) {
		chaincode(node, "defaultchannel", "Asset").GetValue("version").SetScalar("", yaml.StringType)
	},
}}

var chaincodeLanguageUnsupportedOperators = []MutationOperator{{
	RuleID: validate.RuleChaincodeLanguageUnsupported,
	Apply: func(node *yaml.Node) {
		chaincode(node, "defaultchannel", "Asset").GetValue("language").GetValue("name").SetScalar("python", yaml.StringType)
	},
}}
