package microceph

import (
	"fmt"
	"mcloud/pkg/commander"
)

type JoinConfig struct {
	joinToken string
	disk      string
}

// Join makes the node join an existing microceph cluster
func Join(cfg JoinConfig) error {
	// Join microceph cluster
	if _, err := commander.ExecCommand(
		"microceph", "join", cfg.joinToken,
	); 
	err != nil {
		return fmt.Errorf("failed to join microceph cluster: %w", err)
	}

	// Add disk to microceph
	if _, err := commander.ExecCommand(
		"microceph", "disk", "add", cfg.disk,
	); 
	err != nil {
		return fmt.Errorf("failed to add disk: %w", err)
	}

	return nil
}	
