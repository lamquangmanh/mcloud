package microovn

import "mcloud/pkg/commander"

func Bootstrap() (string, error) {
	output, err := commander.ExecCommand("microovn", "init")
	return output, err
}