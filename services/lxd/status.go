package lxd

import "mcloud/pkg/commander"

// ClusterStatus retrieves the status of the LXD cluster
func ClusterStatus() (string, error) {
	return commander.ExecCommand("lxc", "cluster", "list")
}
