package microovn

import "mcloud/pkg/commander"

// Join makes the node join an existing microovn cluster
func Join(token string) (string, error) {
	output, err := commander.ExecCommand("microovn", "join", token)
	return output, err
}