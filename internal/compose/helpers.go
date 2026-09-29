package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/docker/docker/client"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/constants"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/executor"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func ResolvePeerPort(port int) int {
	if port == 0 {
		return constants.DEFAULT_PEER_PORT
	}

	return port
}

func ResolveOrdererPort(port int) int {
	if port == 0 {
		return constants.DEFAULT_ORDERER_PORT
	}

	return port
}

func ResolvePeerVersion(version string) string {
	if version == "" {
		return constants.DEFAULT_FABRIC_VERSION
	}

	if len(strings.Split(version, ".")) == 2 {
		return fmt.Sprintf("%s.0", version)
	}

	return version
}

func ResolvePeerDomain(subdomain string, domain string) string {
	return fmt.Sprintf("%s.%s", subdomain, domain)
}

func ResolveCouchDBDomain(subdomain string, domain string) string {
	peerDomain := ResolvePeerDomain(subdomain, domain)

	return fmt.Sprintf("couchdb.%s", peerDomain)
}

func ResolveOrdererDomain(subdomain string, domain string) string {
	return fmt.Sprintf("%s.%s", subdomain, domain)
}

func ResolveOrdererVersion(version string) string {
	if version == "" {
		return constants.DEFAULT_FABRIC_VERSION
	}

	if len(strings.Split(version, ".")) == 2 {
		return fmt.Sprintf("%s.0", version)
	}

	return version
}

func ResolveCertificateAuthorityDomain(domain string) string {
	return fmt.Sprintf("ca.%s", domain)
}

func ResolveCertificateAuthorityVersion(version string) string {
	if version == "" {
		return "latest"
	}

	return version
}

func ResolveCertificateAuthorityImage(certificateAuthority config.CertificateAuthority) string {
	return fmt.Sprintf("hyperledger/fabric-ca:%s", certificateAuthority.Version)
}

func ResolvePeerImage(peer config.Peer) string {
	return fmt.Sprintf("hyperledger/fabric-peer:%s", ResolvePeerVersion(peer.Version))
}

func ResolveChaincodeCompilerImage(peer config.Peer) string {
	peerVersion := ResolvePeerVersion(peer.Version)

	parts := strings.Split(peerVersion, ".")

	if len(parts) < 2 {
		return fmt.Sprintf("hyperledger/fabric-ccenv:%s", peerVersion)
	}

	return fmt.Sprintf("hyperledger/fabric-ccenv:%s", strings.Join(parts[:2], "."))
}

func ResolveCouchDBImage() string {
	return "couchdb:latest"
}

func ResolveOrdererImage(orderer config.Orderer) string {
	return fmt.Sprintf("hyperledger/fabric-orderer:%s", orderer.Version)
}

func ResolveToolsImage(capabilities config.Capabilities) string {
	if capabilities.Channel == "V3_0" {
		return "ghcr.io/gca-research-group/fabric-tools:3.1.4"
	}

	version := ResolvePeerVersion(config.DefaultVersionByCapability[capabilities.Application])

	return fmt.Sprintf("hyperledger/fabric-tools:%s", version)
}

func ResolveDockerNetworkName(network string) string {
	if network == "" {
		return constants.DEFAULT_NETWORK
	}

	return network
}

func IsDockerRunning() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return false
	}

	defer cli.Close()

	_, err = cli.Ping(context.Background())
	return err == nil
}

func RemoveContainersInNetwork(network string, exec executor.Executor) error {
	out, err := exec.OutputCommand(
		"docker", "network", "inspect", network, "--format",
		"{{ range .Containers }}{{ .Name }}{{ \"\\n\" }}{{ end }}",
	)
	if err != nil {
		return err
	}

	containers := strings.Fields(string(out))

	if len(containers) == 0 {
		return nil
	}

	for _, c := range containers {
		if err := exec.ExecCommand("docker", "rm", "-f", c); err != nil {
			return err
		}
	}

	return nil
}

func ResolvePeerContainerName(domain string, subdomain string) string {
	return fmt.Sprintf("%s.%s", subdomain, domain)
}

func ResolvePeerDockerComposeFile(output string, domain string, subdomain string) string {
	return fmt.Sprintf("%[1]s/%[2]s/%[3]s.yml", output, domain, ResolvePeerDomain(subdomain, domain))
}

func ResolveCertificateAuthorityDockerComposeFile(output string, domain string) string {
	return fmt.Sprintf("%s/%s/%s.yml", output, domain, ResolveCertificateAuthorityDomain(domain))
}

func ResolveOrdererDockerComposeFile(output string, domain string, subdomain string) string {
	return fmt.Sprintf("%[1]s/%[2]s/%[3]s.yml", output, domain, ResolveOrdererDomain(subdomain, domain))
}

func ResolveToolsDockerComposeFile(output string, domain string) string {
	return fmt.Sprintf("%s/%s/tools.yml", output, domain)
}

func ResolveNetworkDockerComposeFile(output string) string {
	return fmt.Sprintf("%s/network.yml", output)
}

func ResolveCertificateAuthorityContainerName(domain string) string {
	return fmt.Sprintf("ca.%s", domain)
}

func ResolveToolsContainerName(organization config.Organization) string {
	return fmt.Sprintf("%s.%s", "tools", organization.Domain)
}

func RunContainerFromTheDockerComposeFile(exec executor.Executor, network string, file string) error {
	args := []string{"compose", "-f", network, "-f", file, "up", "--build", "-d"}
	return exec.ExecCommand("docker", args...)
}

func StopContainerFromTheDockerComposeFile(exec executor.Executor, network string, file string) error {
	args := []string{"compose", "-f", network, "-f", file, "down"}
	return exec.ExecCommand("docker", args...)
}

func ResolveChaincodeHostDir(chaincode config.Chaincode) string {
	currentDir, _ := os.Getwd()

	hostDir := chaincode.Path

	if !filepath.IsAbs(hostDir) {
		hostDir = filepath.Join(currentDir, hostDir)
	}

	return hostDir
}

func configuredChaincodeLanguages(cfg config.Config) mapset.Set[string] {
	languages := mapset.NewSet[string]()
	for _, chaincode := range config.ResolveChaincodes(cfg) {
		languages.Add(chaincode.Language.Name)
	}
	return languages
}

func resolveChaincodeRuntimeImage(peer config.Peer, language string) string {
	version := strings.TrimPrefix(ResolveChaincodeCompilerImage(peer), "hyperledger/fabric-ccenv:")
	if strings.HasPrefix(version, "3.") {
		version = "2.5"
	}

	return "hyperledger/fabric-" + map[string]string{spec.LanguageJava: "javaenv", spec.LanguageNode: "nodeenv"}[language] + ":" + version
}
