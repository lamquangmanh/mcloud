package lxd

import (
	"fmt"
	"mcloud/pkg/commander"
)

type ValidateConfig struct {
	clusterName string
	address     string // IP:8443
}

// Validate checks if the LXD cluster can be initialized with the given configuration
func Validate(cfg ValidateConfig) (bool, error) {
	_, err := commander.ExecCommand(
		"lxd", "init", 
		"--cluster",
		"--cluster-name", cfg.clusterName,
		"--cluster-address", cfg.address, 
		"--auto",
	)
	if err != nil {
		return false, fmt.Errorf("failed to validate LXD cluster: %w", err)
	}

	return true, nil
}
