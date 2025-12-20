package database

import (
	"context"
	"database/sql"
	"time"
)

type Node struct {
	ID            string
	ClusterID     string
	Hostname      string
	IP            string
	Role          string
	Status        string
	JoinedAt      time.Time
	LastHeartbeat *time.Time

	CreatedAt    time.Time
	CreateUserID *string
	UpdatedAt    time.Time
	UpdateUserID *string
}

type NodeRepository struct {
	exec sqlExecutor
}

func NewNodeRepository(db *sql.DB) *NodeRepository {
	return &NodeRepository{exec: db}
}

func NewNodeRepositoryTx(tx *sql.Tx) *NodeRepository {
	return &NodeRepository{exec: tx}
}

func (r *NodeRepository) Create(ctx context.Context, n *Node) error {
	_, err := r.exec.ExecContext(ctx, `
INSERT INTO nodes (
id, cluster_id, hostname, ip, role, status, create_user_id
) VALUES (?, ?, ?, ?, ?, ?, ?)
`, n.ID, n.ClusterID, n.Hostname, n.IP, n.Role, n.Status, n.CreateUserID)
	return err
}

func (r *NodeRepository) UpdateByID(ctx context.Context, n *Node) error {
	_, err := r.exec.ExecContext(ctx, `
UPDATE nodes
SET hostname = ?, ip = ?, role = ?, status = ?,
updated_at = CURRENT_TIMESTAMP, update_user_id = ?
WHERE id = ?
`, n.Hostname, n.IP, n.Role, n.Status, n.UpdateUserID, n.ID)
	return err
}

func (r *NodeRepository) UpdateHeartbeat(ctx context.Context, nodeID string) error {
	_, err := r.exec.ExecContext(ctx, `
UPDATE nodes SET last_heartbeat = CURRENT_TIMESTAMP WHERE id = ?
`, nodeID)
	return err
}

func (r *NodeRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	return err
}

func (r *NodeRepository) GetByID(ctx context.Context, id string) (*Node, error) {
	row := r.exec.QueryRowContext(ctx, `
SELECT id, cluster_id, hostname, ip, role, status,
joined_at, last_heartbeat,
created_at, create_user_id, updated_at, update_user_id
FROM nodes WHERE id = ?
`, id)

	var n Node
	if err := row.Scan(
		&n.ID, &n.ClusterID, &n.Hostname, &n.IP,
		&n.Role, &n.Status, &n.JoinedAt, &n.LastHeartbeat,
		&n.CreatedAt, &n.CreateUserID, &n.UpdatedAt, &n.UpdateUserID,
	); err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NodeRepository) ListByCluster(ctx context.Context, clusterID string) ([]Node, error) {
	rows, err := r.exec.QueryContext(ctx, `
SELECT id, cluster_id, hostname, ip, role, status,
joined_at, last_heartbeat,
created_at, create_user_id, updated_at, update_user_id
FROM nodes WHERE cluster_id = ?
`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(
			&n.ID, &n.ClusterID, &n.Hostname, &n.IP,
			&n.Role, &n.Status, &n.JoinedAt, &n.LastHeartbeat,
			&n.CreatedAt, &n.CreateUserID, &n.UpdatedAt, &n.UpdateUserID,
		); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, nil
}
