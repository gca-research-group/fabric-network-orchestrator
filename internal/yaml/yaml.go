package yaml

import (
	"bytes"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Node yaml.Node

type ScalarType string

const (
	StringType ScalarType = "string"
	IntType    ScalarType = "int"
	FloatType  ScalarType = "float"
	BoolType   ScalarType = "bool"
)

func FromBytes(data []byte) (*Node, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)

	return (*Node)(&root), err
}

func (n *Node) SetScalar(value string, scalarType ScalarType) *Node {
	n.Kind = yaml.ScalarNode
	n.Value = value
	n.Style = 0

	switch scalarType {
	case StringType:
		n.Tag = "!!str"
		n.Style = yaml.DoubleQuotedStyle

	case IntType:
		n.Tag = "!!int"

	case FloatType:
		n.Tag = "!!float"

	case BoolType:
		n.Tag = "!!bool"
	}

	return n
}

func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}

	clone := *n

	clone.Content = make(
		[]*yaml.Node,
		len(n.Content),
	)

	for i, child := range n.Content {
		clonedChild := (*Node)(child).Clone()
		clone.Content[i] = (*yaml.Node)(clonedChild)
	}

	if n.Alias != nil {
		clone.Alias = (*yaml.Node)(
			(*Node)(n.Alias).Clone(),
		)
	}

	return &clone
}

func (n *Node) FindByValue(
	key string,
	value string,
) *Node {

	for _, content := range n.Content {
		item := (*Node)(content)

		if item.Kind != yaml.MappingNode {
			continue
		}

		property := item.GetValue(key)

		if property != nil && property.Value == value {
			return item
		}
	}

	return nil
}

func (n *Node) Document() *Node {
	if n == nil || len(n.Content) == 0 {
		return nil
	}

	return (*Node)(n.Content[0])
}

func ScalarNode(value string) *Node {
	return (*Node)(&yaml.Node{Kind: yaml.ScalarNode, Value: value})
}

func MappingNode(content ...*Node) *Node {
	_content := make([]*yaml.Node, len(content))
	for i, c := range content {
		_content[i] = (*yaml.Node)(c)
	}

	return (*Node)(&yaml.Node{Kind: yaml.MappingNode, Content: _content})
}

func SequenceNode(content ...*Node) *Node {
	_content := make([]*yaml.Node, len(content))
	for i, c := range content {
		_content[i] = (*yaml.Node)(c)
	}

	return (*Node)(&yaml.Node{Kind: yaml.SequenceNode, Content: _content})
}

func AliasNode(value string, alias *Node) *Node {
	return (*Node)(&yaml.Node{Kind: yaml.AliasNode, Value: value, Alias: (*yaml.Node)(alias)})
}

func (n *Node) WithAnchor(name string) *Node {
	n.Anchor = name
	return n
}

func (n *Node) WithTag(name string) *Node {
	n.Tag = name
	return n
}

func (n *Node) MarshalYAML() (*yaml.Node, error) {
	if n == nil {
		return nil, nil
	}
	return (*yaml.Node)(n), nil
}

func (n *Node) GetValue(keyName string) *Node {
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == keyName {
			return (*Node)(n.Content[i+1])
		}
	}

	return nil
}

func (n *Node) GetOrCreateValue(keyName string, defaultNode *Node) *Node {
	node := n.GetValue(keyName)

	if node != nil {
		return node
	}

	n.Content = append(n.Content, (*yaml.Node)(ScalarNode(keyName)), (*yaml.Node)(defaultNode))

	return defaultNode
}

func (n *Node) WithStyle(style yaml.Style) *Node {
	n.Style = style
	return n
}

func (n *Node) WithDoubleQuotedStyle() *Node {
	n.WithStyle(yaml.DoubleQuotedStyle)
	return n
}

func (n *Node) WithFolderStyle() *Node {
	n.WithStyle(yaml.FoldedStyle)
	return n
}

func (n *Node) WithFlowStyle() *Node {
	n.WithStyle(yaml.FlowStyle)
	return n
}

func (n *Node) WithLiteralStyle() *Node {
	n.WithStyle(yaml.LiteralStyle)
	return n
}

func (n *Node) ToFile(name string) error {
	_, err := os.Stat(filepath.Dir(name))

	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(name), 0755)
	} else if err != nil {
		return err
	}

	node, err := n.MarshalYAML()

	if err != nil {
		return err
	}

	f, err := os.Create(name)
	if err != nil {
		return err
	}

	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)

	if err := encoder.Encode(node); err != nil {
		return err
	}

	return nil
}

// ToBytes encodes the YAML document using the same formatting as ToFile.
func (n *Node) ToBytes() ([]byte, error) {
	node, err := n.MarshalYAML()
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
