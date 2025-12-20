package lxd

// MockClient is a mock implementation of the LXD Client interface for testing
type MockClient struct {
	InitClusterFunc func(address string) (*NodeInfo, error)
}

func (m *MockClient) InitCluster(address string) (*NodeInfo, error) {
	if m.InitClusterFunc != nil {
		return m.InitClusterFunc(address)
	}
	// Default mock behavior
	return &NodeInfo{
		Hostname: "mock-node",
		IP:       address,
	}, nil
}

// NewMockClient creates a new mock LXD client for testing
func NewMockClient() *MockClient {
	return &MockClient{}
}
