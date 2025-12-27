package lxd

import (
	"fmt"
)

type JoinConfig struct {
	nodeName		 string
	nodeAddress	 string
	clusterAddress string
	clusterCertificate string
	clusterToken   string
}

// generateJoinConfig creates the init config YAML for joining an LXD cluster
func generateJoinConfig(
	nodeName string,
	nodeAddress string,
	leaderAddress string,
	clusterCert string,
	clusterToken string,
) (*InitConfigYaml, error) {
	return &InitConfigYaml{
		Config: map[string]string{
			"core.https_address": nodeAddress + ":8443",
		},
		Cluster: ClusterConfigYaml{
			Enabled:            true,
			ServerName:         nodeName,
			ClusterAddress:     leaderAddress + ":8443",
			ClusterCertificate: clusterCert,
			ClusterToken:       clusterToken,
		},
	}, nil
}

// JoinCluster joins an existing LXD cluster with the given configuration
func JoinCluster(cfg JoinConfig) (string, error) {
	// generate init config
	data, err := generateJoinConfig(cfg.nodeName, cfg.nodeAddress, cfg.clusterAddress, cfg.clusterCertificate, cfg.clusterToken)
	if err != nil {
		return "", fmt.Errorf("failed to generate init config: %w", err)
	}

	// run lxd init with preseed
	initErr := RunInit(data)
	if initErr != nil {
		return "", fmt.Errorf("failed to join LXD cluster: %w", initErr)
	}

	return "LXD cluster joined successfully", nil
}
