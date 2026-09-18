package compose

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/executor"

	mapset "github.com/deckarep/golang-set/v2"
)

func PullChaincodeCompilerImages(config config.Config, executor executor.Executor) error {
	images := mapset.NewSet[string]()

	for _, organization := range config.Organizations {
		for _, peer := range organization.Peers {
			images.Add(ResolveChaincodeCompilerImage(peer))
		}
	}

	orderedImages := images.ToSlice()
	sort.Strings(orderedImages)
	for _, image := range orderedImages {
		output, err := executor.OutputCommand("docker", "image", "ls", "--quiet", "--filter", "reference="+image)
		if err != nil {
			return fmt.Errorf("check compiler image %s: %w", image, err)
		}
		if strings.TrimSpace(string(output)) != "" {
			continue
		}
		if _, err := executor.OutputCommand("docker", "pull", image); err != nil {
			return fmt.Errorf("pull compiler image %s: %w", image, err)
		}
	}

	return nil
}

func PullImages(config config.Config, executor executor.Executor) error {

	images := mapset.NewSet[string]()
	images.Add(ResolveCouchDBImage())

	for _, organization := range config.Organizations {
		images.Add(ResolveCertificateAuthorityImage(organization.CertificateAuthority))

		for _, peer := range organization.Peers {
			images.Add(ResolvePeerImage(peer))
			images.Add(ResolveChaincodeCompilerImage(peer))
		}

		for _, orderer := range organization.Orderers {
			images.Add(ResolveOrdererImage(orderer))
		}

		images.Add(ResolveToolsImage(config.Capabilities))
	}

	for image := range images.Iter() {
		if _, err := executor.OutputCommand("docker", "pull", image); err != nil {
			return fmt.Errorf("Error when pulling the image %s: %v\n", image, err)
		}
	}

	return nil
}
