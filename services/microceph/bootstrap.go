package microceph

import (
	"mcloud/pkg/commander"
	"mcloud/pkg/logger"
)

type BootstrapConfig struct {
	Disk string // example: /dev/sdb
}

// Bootstrap initializes the microceph service with the given configuration
func Bootstrap(cfg BootstrapConfig) error {
	// Initialize microceph
	if _, err := commander.ExecCommand("microceph", "init"); 
	err != nil {
		logger.Error("failed to init microceph: %v", err)
		return err
	}

	// Add disk to microceph
	if _, err := commander.ExecCommand(
		"microceph", "disk", "add", cfg.Disk,
	); 
	err != nil {
		logger.Error("failed to add disk: %v", err)
		return err
	}

	return nil
}
