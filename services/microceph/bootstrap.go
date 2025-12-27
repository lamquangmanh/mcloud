package microceph

import (
	"fmt"
	"mcloud/pkg/commander"
)

type BootstrapConfig struct {
	disk string // example: /dev/sdb
}

// Bootstrap initializes the microceph service with the given configuration
func Bootstrap(cfg BootstrapConfig) error {
	// Initialize microceph
	if _, err := commander.ExecCommand("microceph", "init"); 
	err != nil {
		fmt.Errorf("failed to init microceph: %w", err)
		return err
	}

	// Add disk to microceph
	if _, err := commander.ExecCommand(
		"microceph", "disk", "add", cfg.disk,
	); 
	err != nil {
		fmt.Errorf("failed to add disk: %w", err)
		return err
	}

	return nil
}
