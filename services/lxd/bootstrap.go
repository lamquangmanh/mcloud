package lxd

import (
	"bytes"
	"fmt"
	"os/exec"

	"gopkg.in/yaml.v3"
)

type InitConfigYaml struct {
	Config   map[string]string `yaml:"config,omitempty"`
	Cluster  ClusterConfigYaml `yaml:"cluster"`
}

type ClusterConfigYaml struct {
	Enabled            bool   `yaml:"enabled"`
	ServerName         string `yaml:"server_name"`
	ClusterAddress     string `yaml:"cluster_address"`
	ClusterCertificate string `yaml:"cluster_certificate,omitempty"`
	ClusterToken       string `yaml:"cluster_token,omitempty"`
}

type BootstrapConfig struct {
	clusterName string
	address     string // IP:8443
}

// generateInitConfig creates the LXD init preseed configuration for bootstrapping a cluster
func generateInitConfig(nodeName string, address string) (*InitConfigYaml, error) {
	return &InitConfigYaml{
		Config: map[string]string{
			"core.https_address": address + ":8443",
		},
		Cluster: ClusterConfigYaml{
			Enabled:        true,
			ServerName:     nodeName,
			ClusterAddress: address + ":8443",
		},
	}, nil
}

// RunInit executes the 'lxd init' command with the provided preseed configuration
func RunInit(initCfg *InitConfigYaml) error {
	data, err := yaml.Marshal(initCfg)
	if err != nil {
		return err
	}

	cmd := exec.Command("lxd", "init", "--preseed")
	cmd.Stdin = bytes.NewReader(data)

	return cmd.Run()
}

// Bootstrap initializes a new LXD cluster with the given configuration
func Bootstrap(cfg BootstrapConfig) (string, error) {
	// generate init config
	data, err := generateInitConfig(cfg.clusterName, cfg.address)
	if err != nil {
		return "", fmt.Errorf("failed to generate init config: %w", err)
	}

	// run lxd init with preseed
	initErr := RunInit(data)
	if initErr != nil {
		return "", fmt.Errorf("failed to bootstrap LXD cluster: %w", initErr)
	}

	return "LXD cluster bootstrapped successfully", nil
}


