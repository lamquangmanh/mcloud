# mcloud

MCloud (ManhLam Cloud).
The goal of this project is to learn cloud computing in depth by building a small private cloud for testing and experimentation.

# Prerequire

Device:

- cpu: 4 cores
- ram: 8Gb
- disk: 10Gb

OS: Ubuntu 22 LTS or higher

Technical:

- Language: Golang
- SQLite + Raft (Dqlite / etcd-lite) [Link][https://canonical.com/dqlite/docs]
- Community: gRPC + mTLS
- LXD: Container/Virtual Machine [Link][https://documentation.ubuntu.com/microcloud/latest/lxd/]
- Network: MicroOVN [Link][https://canonical-microovn.readthedocs-hosted.com/en/latest/]
- Storage: MicroCeph [Link][https://canonical-microceph.readthedocs-hosted.com/stable/]

# Features:

- Node Management: Init cluster, join node, leave node.
- VPC (Virtual Private Cloud): Subnet public, Subnet private, Route table, Internet Gateway, NAT Gateway, Firewall
-
