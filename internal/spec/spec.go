package spec

const (
	LanguageGo   = "golang"
	LanguageJava = "java"
	LanguageNode = "node"
)

type CapabilityLevel int

const (
	V2_0 CapabilityLevel = iota + 1
	V2_5
	V3_0
)

var CapabilityMap = map[string]CapabilityLevel{
	"V2_0": V2_0,
	"V2_5": V2_5,
	"V3_0": V3_0,
}

var MinBinaryVersion = map[string]string{
	"V2_0": "2.0.0",
	"V2_5": "2.5.0",
	"V3_0": "3.0.0",
}

var DefaultVersionByCapability = map[string]string{
	"V2_0": "2.5.15",
	"V2_5": "2.5.15",
	"V3_0": "3.1.4",
}

type Orderer struct {
	Name       string `yaml:"name" json:"name" toml:"name"`
	Subdomain  string `yaml:"subdomain" json:"subdomain" toml:"subdomain"`
	Port       int    `yaml:"port" json:"port" toml:"port"`
	ExposePort int    `yaml:"exposePort" json:"exposePort" toml:"exposePort"`
	Version    string `yaml:"version" json:"version" toml:"version"`
}

type Peer struct {
	Name       string `yaml:"name" json:"name" toml:"name"`
	Subdomain  string `yaml:"subdomain" json:"subdomain" toml:"subdomain"`
	Port       int    `yaml:"port" json:"port" toml:"port"`
	ExposePort int    `yaml:"exposePort" json:"exposePort" toml:"exposePort"`
	Version    string `yaml:"version" json:"version" toml:"version"`
	IsAnchor   bool   `yaml:"isAnchor" json:"isAnchor" toml:"isAnchor"`
}

type CertificateAuthority struct {
	ExposePort int    `yaml:"exposePort,omitempty" json:"exposePort,omitempty" toml:"bootstrap,omitempty"`
	Version    string `yaml:"version" json:"version" toml:"version"`
}

type Organization struct {
	Name                 string               `yaml:"name" json:"name" toml:"name"`
	Domain               string               `yaml:"domain" json:"domain" toml:"domain"`
	Peers                []Peer               `yaml:"peers" json:"peers" toml:"peers"`
	Orderers             []Orderer            `yaml:"orderers" json:"orderers" toml:"orderers"`
	Users                int                  `yaml:"users" json:"users" toml:"users"`
	CertificateAuthority CertificateAuthority `yaml:"certificateAuthority" json:"certificateAuthority" toml:"certificateAuthority"`
	Bootstrap            bool                 `yaml:"bootstrap,omitempty" json:"bootstrap,omitempty" toml:"bootstrap,omitempty"`
}

type Channel struct {
	Name       string      `yaml:"name" json:"name" toml:"name"`
	Profile    string      `yaml:"profile" json:"profile" toml:"profile"`
	Chaincodes []Chaincode `yaml:"chaincodes" json:"chaincodes" toml:"chaincodes"`
}

type Consensus struct {
	Type string `yaml:"type" json:"type" toml:"type"`
}

type Profile struct {
	Name          string    `yaml:"name" json:"name" toml:"name"`
	Organizations []string  `yaml:"organizations" json:"organizations" toml:"organizations"`
	Consensus     Consensus `yaml:"consensus" json:"consensus" toml:"consensus"`
}

type Capabilities struct {
	Channel     string `yaml:"channel" json:"channel" toml:"channel"`
	Orderer     string `yaml:"orderer" json:"orderer" toml:"orderer"`
	Application string `yaml:"application" json:"application" toml:"application"`
}

type Language struct {
	Name string `yaml:"name" json:"name" toml:"name"`
}

type Chaincode struct {
	Path                string   `yaml:"path" json:"path" toml:"path"`
	Name                string   `yaml:"name" json:"name" toml:"name"`
	Version             string   `yaml:"version" json:"version" toml:"version"`
	Language            Language `yaml:"language" json:"language" toml:"language"`
	SignaturePolicy     string   `yaml:"signaturePolicy" json:"signaturePolicy" toml:"signaturePolicy"`
	ChannelConfigPolicy string   `yaml:"channelConfigPolicy" json:"channelConfigPolicy" toml:"channelConfigPolicy"`
	CollectionsConfig   string   `yaml:"collectionsConfig" json:"collectionsConfig" toml:"collectionsConfig"`
}

type Config struct {
	Output        string         `yaml:"output" json:"output" toml:"output"`
	Network       string         `yaml:"network" json:"network" toml:"network"`
	Capabilities  Capabilities   `yaml:"capabilities" json:"capabilities" toml:"capabilities"`
	Organizations []Organization `yaml:"organizations" json:"organizations" toml:"organizations"`
	Profiles      []Profile      `yaml:"profiles" json:"profiles" toml:"profiles"`
	Channels      []Channel      `yaml:"channels" json:"channels" toml:"channels"`
}
