package chaincode

import (
	"fmt"
	"log/slog"

	"github.com/gca-research-group/fabric-network-orchestrator/internal/config"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
)

func (c *Chaincode) Package() error {

	organization := c.config.Organizations[0]
	chaincodes := config.ResolveChaincodes(*c.config)

	for _, chaincode := range chaincodes {

		name := chaincode.Name
		label := ResolveLabel(chaincode)
		tarfile := ResolveChaincodeTar(chaincode)
		chaincodePath := ResolveChaincodePath(chaincode)
		language := chaincode.Language.Name

		type packageStep struct {
			name    string
			message string
			args    []string
		}

		steps := []packageStep{}

		if language == spec.LanguageGo {
			steps = append(steps, packageStep{"Initialize", "Error when initializing the chaincode module %s: %v", []string{
				"sh", "-c", fmt.Sprintf("cd %s && [ -f go.mod ] || go mod init %s; go mod tidy", chaincodePath, name),
			}})
		}

		steps = append(steps, packageStep{"Package", "Error when packaging the chaincode %s: %v", []string{
			"peer", "lifecycle", "chaincode", "package", tarfile,
			"--path", chaincodePath,
			"--lang", language,
			"--label", label,
		}})

		for _, step := range steps {
			slog.Info("Executing step", "step", step.name)
			_, err := c.ExecInTools(organization, step.args)
			if err != nil {
				return fmt.Errorf(step.message, name, err)
			}
		}
	}

	return nil
}
