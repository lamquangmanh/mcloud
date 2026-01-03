package cluster

import (
	"database/sql"
	"errors"

	"mcloud/internal/database"
	"mcloud/pkg/commander"
	// "mcloud/services/lxd"
)

const (
	Disk = "/dev/sdb" // disk used for microceph
)

type Service struct {
	db        *sql.DB
	// lxdClient lxd.Client
}

type InitRequest struct {
	Name             string `json:"name"`
	AdvertiseAddress string `json:"advertise_address"`
}

type InitResult struct {
	ClusterID string         `json:"cluster_id"`
	Token     string         `json:"token"`
	Leader    *database.Node `json:"leader"`
}

func NewService(db *sql.DB) *Service {
	// Create LXD client
	// lxdClient := lxd.NewClient()
	return &Service{
		db:        db,
		// lxdClient: lxdClient,
	}
}

func validateInitRequest(req *InitRequest) error {
	// Basic validation
	if req.Name == "" {
		return errors.New("cluster name is required")
	}
	if req.AdvertiseAddress == "" {
		return errors.New("advertise address is required")
	}

	// check snap lxd, microceph, microovn installed
	cmds := []string{
		"lxd",
		"lxc",
		"microceph",
		"microovn",
	}
	for _, c := range cmds {
		if err := commander.CheckCommandExists(c); err != nil {
			return err
		}
	}

	// check port 8443 available
	if err := commander.CheckPortAvailable(8443); err != nil {
		return err
	}

	// check disk exists
	if err := commander.CheckDiskExists(Disk); err != nil {
		return err
	}
	
	return nil
}

// func (s *Service) InitCluster(ctx context.Context, req *InitRequest) (*InitResult, error) {
// 	// 1. Validate
// 	if err := validateInitRequest(req); err != nil {
// 		return nil, err
// 	}

// 	// 2. Check cluster exists (fast-fail)
// 	clusterRepo := database.NewClusterRepository(s.db)
// 	count, err := clusterRepo.Count(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if count > 0 {
// 		return nil, errors.New("cluster already initialized")
// 	}

// 	// 3. Generate data (NO DB)
// 	clusterID := uuid.NewString()

// 	// Generate CA certificate
// 	caCertPEM, caKeyPEM, err := cert.GenerateCA("", "MCloud Cluster CA")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Generate bootstrap token
// 	token := auth.GenerateJoinToken(clusterID)
// 	tokenExpiry := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

// 	// 4. LXD INIT (SIDE EFFECT)
// 	// nodeInfo, err := s.lxdClient.InitCluster(req.AdvertiseAddress)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	var result *InitResult

// 	// 5. Persist state (TRANSACTION ONLY)
// 	err = database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
// 		clusterRepo := database.NewClusterRepositoryTx(tx)
// 		nodeRepo := database.NewNodeRepositoryTx(tx)
// 		caRepo := database.NewCertificateAuthorityRepositoryTx(tx)
// 		tokenRepo := database.NewBootstrapTokenRepositoryTx(tx)

// 		cluster := &database.Cluster{
// 			ID:    clusterID,
// 			Name:  req.Name,
// 			State: "active",
// 		}

// 		if err := clusterRepo.Create(ctx, cluster); err != nil {
// 			return err
// 		}

// 		node := &database.Node{
// 			ID:        uuid.NewString(),
// 			ClusterID: clusterID,
// 			Hostname:  nodeInfo.Hostname,
// 			IP:        nodeInfo.IP,
// 			Role:      "leader",
// 			Status:    "online",
// 		}

// 		if err := nodeRepo.Create(ctx, node); err != nil {
// 			return err
// 		}

// 		ca := &database.CertificateAuthority{
// 			ID:        uuid.NewString(),
// 			ClusterID: clusterID,
// 			CertPEM:   caCertPEM,
// 			KeyPEM:    caKeyPEM,
// 		}

// 		if err := caRepo.Create(ctx, ca); err != nil {
// 			return err
// 		}

// 		bootstrapToken := &database.BootstrapToken{
// 			Token:     token,
// 			ClusterID: clusterID,
// 			ExpiresAt: tokenExpiry,
// 			Used:      false,
// 		}

// 		if err := tokenRepo.Create(ctx, bootstrapToken); err != nil {
// 			return err
// 		}

// 		// Store LXD, Ceph, and OVN configurations
// 		kvRepo := database.NewKVStoreRepositoryTx(tx)
		
// 		// Store LXD cluster configuration
// 		if err := kvRepo.Set(ctx, "lxd.cluster.name", req.Name); err != nil {
// 			return err
// 		}
// 		if err := kvRepo.Set(ctx, "lxd.cluster.address", req.AdvertiseAddress); err != nil {
// 			return err
// 		}
		
// 		// Store Ceph configuration placeholders
// 		if err := kvRepo.Set(ctx, "ceph.enabled", "true"); err != nil {
// 			return err
// 		}
// 		if err := kvRepo.Set(ctx, "ceph.cluster.name", req.Name+"-ceph"); err != nil {
// 			return err
// 		}
		
// 		// Store OVN configuration placeholders
// 		if err := kvRepo.Set(ctx, "ovn.enabled", "true"); err != nil {
// 			return err
// 		}
// 		if err := kvRepo.Set(ctx, "ovn.network.name", req.Name+"-ovn"); err != nil {
// 			return err
// 		}

// 		result = &InitResult{
// 			ClusterID: clusterID,
// 			Leader:    node,
// 			Token:     token,
// 		}
// 		return nil
// 	})

// 	return result, err
// }

// func (s *Service) InitCluster(ctx context.Context, req *InitRequest) (*InitResult, error) {
// 	var result *InitResult

// 	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
// 		clusterRepo := database.NewClusterRepositoryTx(tx)
		
// 		// Check if a cluster already exists
// 		count, err := clusterRepo.Count(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		if count > 0 {
// 			return errors.New("cluster already initialized")
// 		}

// 		caRepo := database.NewCARepositoryTx(tx)
// 		tokenRepo := database.NewTokenRepositoryTx(tx)

// 		cluster := &database.Cluster{
// 			ID:    uuid.NewString(),
// 			Name:  req.Name,
// 			State: "active",
// 		}

// 		if err := clusterRepo.Create(ctx, cluster); err != nil {
// 			return err
// 		}

// 		ca, err := s.certService.CreateCA(cluster.ID)
// 		if err != nil {
// 			return err
// 		}
// 		if err := caRepo.Create(ctx, ca); err != nil {
// 			return err
// 		}

// 		token := auth.NewBootstrapToken(cluster.ID)
// 		if err := tokenRepo.Create(ctx, token); err != nil {
// 			return err
// 		}

// 		result = &InitResult{
// 			ClusterID: cluster.ID,
// 			// Token:     "",
// 			Token:     token.Token,
// 		}
// 		return nil
// 	})

// 	return result, err
// }


// func (s *Service) InitCluster(ctx context.Context, req *InitRequest) (*InitResult, error) {
// 	repo := store.NewClusterRepository(s.db)
// 	count, err := repo.Count(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if count > 0 {
// 		return nil, errors.New("cluster already initialized")
// 	}

// 	clusterID := uuid.NewString()
// 	token := "mcloud-bootstrap-token"

// 	err = store.WithTx(ctx, s.db, func(tx *sql.Tx) error {
// 		repo := store.NewClusterRepositoryTx(tx)
// 		return repo.Create(ctx, &store.Cluster{
// 			ID:    clusterID,
// 			Name:  req.Name,
// 			State: "active",
// 		})
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &InitResult{
// 		ClusterID: clusterID,
// 		Token:     token,
// 	}, nil
// }