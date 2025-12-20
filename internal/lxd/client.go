package lxd

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

type InitClusterRequest struct {
	ClusterName  string
	Address      string // IP:8443
	ClusterToken string
}

type NodeInfo struct {
	Hostname string
	IP       string
}

type Client interface {
	InitCluster(address string) (*NodeInfo, error)
}

type LxdClient struct {
	socketPath string
}

func NewClient() Client {
	return &LxdClient{
		socketPath: "/var/snap/lxd/common/lxd/unix.socket",
	}
}

func NewLxdClient() *LxdClient {
	return &LxdClient{
		socketPath: "/var/snap/lxd/common/lxd/unix.socket",
	}
}

func (c *LxdClient) httpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", c.socketPath)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}