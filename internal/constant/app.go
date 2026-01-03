package constant

const (
	// AppName is the name of the application
	AppName = "mcloud"
	AppServerName = "mcloud-server"

	// AppVersion is the current version of the application
	AppVersion = "0.1.0"

	// OrganizationName is the name of the organization
	OrganizationName = "MCloud"

	// RootCACommonName is the common name for the root CA certificate
	RootCACommonName = "MCloud Cluster CA"

	DefaultConfigPath = "/etc/mcloud/config.yaml"
	DefaultStatePath  = "/var/lib/mcloud/state.yaml"
)

type NodeRole string

const (
	RoleLeader NodeRole = "leader"
	RoleMember NodeRole = "member"
)

const (
	// DefaultCephDisk is the default disk for mcloud
	DefaultCephDisk = "/dev/sdb"
)