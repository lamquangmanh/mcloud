package microceph

import "mcloud/pkg/commander"

// RegisterToLXD registers the given Ceph pool to LXD
func RegisterToLXD(pool string) (string, error) {
	output, err := commander.ExecCommand(
		"lxc", "storage", "create",
		"ceph-"+pool,
		"ceph",
		"source="+pool,
	)
	return output, err
}
