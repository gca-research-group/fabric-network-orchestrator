package application

import (
	"fmt"
	"log/slog"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/chaincode"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/compose"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/configtx"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/directory"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/executor"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/network"
)

type Workflows struct {
	executor executor.Executor
}

func NewWorkflows(exec executor.Executor) *Workflows {
	if exec == nil {
		exec = &executor.DefaultExecutor{}
	}
	return &Workflows{executor: exec}
}

// Deploy generates artifacts and deploys the network and configured chaincodes.
// On failure, completed stages are preserved for inspection and recovery.
func (w *Workflows) Deploy(cfg *config.Config) error {
	steps := []struct {
		name string
		fn   func(*config.Config) error
	}{
		{"generate artifacts", w.GenerateArtifacts},
		{"deploy network", w.DeployNetwork},
		{"deploy chaincodes", w.DeployChaincodes},
	}
	for _, step := range steps {
		slog.Info("Executing deployment stage", "stage", step.name)
		if err := step.fn(cfg); err != nil {
			return fmt.Errorf("deployment failed at stage %s: %w", step.name, err)
		}
	}
	return nil
}

func (w *Workflows) GenerateArtifacts(cfg *config.Config) error {
	empty, err := directory.IsDirEmpty(cfg.Output)
	if err != nil {
		return fmt.Errorf("inspect artifact directory: %w", err)
	}
	if !empty {
		return fmt.Errorf("artifact directory is not empty: %s", cfg.Output)
	}
	if err := configtx.NewRenderer(cfg).Render(); err != nil {
		return fmt.Errorf("render configtx artifacts: %w", err)
	}
	if err := compose.NewRenderer(cfg).Render(); err != nil {
		return fmt.Errorf("render Docker Compose artifacts: %w", err)
	}
	return nil
}

func (w *Workflows) CleanArtifacts(cfg *config.Config) error {
	if err := directory.RemoveFolderIfExists(cfg.Output); err != nil {
		return fmt.Errorf("clean artifact directory: %w", err)
	}
	return nil
}

func (w *Workflows) GenerateIdentities(cfg *config.Config) error {
	containers := network.NewContainerManager(*cfg, w.executor)
	if err := containers.RunCAContainers(); err != nil {
		return fmt.Errorf("start certificate authorities: %w", err)
	}
	if err := network.NewIdentityManager(*cfg, w.executor).GenerateAll(); err != nil {
		return fmt.Errorf("generate identities: %w", err)
	}
	if err := containers.StopCertificateAuthorities(); err != nil {
		return fmt.Errorf("stop certificate authorities: %w", err)
	}
	return nil
}

func (w *Workflows) PullImages(cfg *config.Config) error {
	if err := compose.PullImages(*cfg, w.executor); err != nil {
		return fmt.Errorf("pull images: %w", err)
	}
	return nil
}

func (w *Workflows) DeployNetwork(cfg *config.Config) error {
	if err := network.NewNetwork(*cfg, w.executor).Deploy(); err != nil {
		return fmt.Errorf("deploy network: %w", err)
	}
	return nil
}

func (w *Workflows) StartNetwork(cfg *config.Config) error {
	if err := network.NewContainerManager(*cfg, w.executor).Start(); err != nil {
		return fmt.Errorf("start network: %w", err)
	}
	return nil
}

func (w *Workflows) StopNetwork(cfg *config.Config) error {
	name := compose.ResolveDockerNetworkName(cfg.Network)
	if err := compose.RemoveContainersInNetwork(name, w.executor); err != nil {
		return fmt.Errorf("stop network: %w", err)
	}
	return nil
}

func (w *Workflows) DeployChaincodes(cfg *config.Config) error {
	if err := compose.PullChaincodeCompilerImages(*cfg, w.executor); err != nil {
		return fmt.Errorf("deploy chaincodes: prepare compiler images: %w", err)
	}

	if err := chaincode.NewChaincode(cfg, w.executor).Publish(); err != nil {
		return fmt.Errorf("deploy chaincodes: %w", err)
	}
	return nil
}
