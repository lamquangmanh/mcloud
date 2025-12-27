package microovn

import "mcloud/pkg/commander"

// RegisterToLXD registers the given OVN network to LXD
func RegisterToLXD(network string) (string, error) {
	output, err := commander.ExecCommand(
		"lxc", "network", "create",
		network,
		"--type=ovn",
	)
	return output, err
}