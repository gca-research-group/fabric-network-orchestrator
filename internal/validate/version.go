package validate

import (
	"fmt"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/spec"
	"strings"
)

func minimumBinaryVersion(capabilities ...string) string {
	var strictest string
	for _, capability := range capabilities {
		if spec.CapabilityMap[capability] > spec.CapabilityMap[strictest] {
			strictest = capability
		}
	}

	return spec.MinBinaryVersion[strictest]
}

func validateBinary(version, minimum string) error {
	if version == "" {
		return nil
	}

	currentParts, minimumParts := parseVersion(version), parseVersion(minimum)

	for i := 0; i < 3; i++ {
		if currentParts[i] > minimumParts[i] {
			return nil
		}

		if currentParts[i] < minimumParts[i] {
			return fmt.Errorf("version %s is lower than required %s", version, minimum)
		}
	}
	return nil
}

func parseVersion(version string) [3]int {
	var result [3]int
	parts := strings.Split(version, ".")

	for i := 0; i < len(parts) && i < 3; i++ {
		fmt.Sscanf(parts[i], "%d", &result[i])
	}

	return result
}
