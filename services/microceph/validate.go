package microceph

import (
	"fmt"
	"mcloud/pkg/commander"
)

type ValidateConfig struct {
	ClusterName string
	Address     string // IP:8443
}

// validate checks if the microceph cluster can be initialized with the given configuration
func Validate(cfg ValidateConfig) (bool, error) {
	_, err := commander.ExecCommand(
		"microceph", "init", 
		"--cluster",
		"--cluster-name", cfg.ClusterName,
		"--cluster-address", cfg.Address, 
		"--auto",
	)
	if err != nil {
		return false, fmt.Errorf("failed to validate microceph cluster: %w", err)
	}

	return true, nil
}
