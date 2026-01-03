package microovn

import (
	"mcloud/pkg/commander"
	"mcloud/pkg/logger"
)

func Bootstrap() error {
	_, err := commander.ExecCommand("microovn", "init")
	if err != nil {
		logger.Error("failed to init microovn: %v", err)
	}
	
	return nil
}