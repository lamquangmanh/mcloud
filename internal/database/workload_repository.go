package database

import (
	"context"
	"database/sql"
	"time"
)

type Workload struct {
	ID           string
	ClusterID    string
	NodeID       *string
	Name         string
	Kind         string
	Status       string
	CreatedAt    time.Time
	CreateUserID *string
	UpdatedAt    time.Time
	UpdateUserID *string
}

type WorkloadRepository struct {
	db *sql.DB
}

func NewWorkloadRepository(db *sql.DB) *WorkloadRepository {
	return &WorkloadRepository{db: db}
}

func (r *WorkloadRepository) Create(ctx context.Context, w *Workload) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO workloads (id, cluster_id, node_id, name, kind, status, create_user_id)
VALUES (?, ?, ?, ?, ?, ?, ?)
`, w.ID, w.ClusterID, w.NodeID, w.Name, w.Kind, w.Status, w.CreateUserID)
	return err
}

func (r *WorkloadRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE workloads
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, status, id)
	return err
}

func (r *WorkloadRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM workloads WHERE id = ?`, id)
	return err
}

func (r *WorkloadRepository) GetByID(ctx context.Context, id string) (*Workload, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, cluster_id, node_id, name, kind, status,
created_at, create_user_id, updated_at, update_user_id
FROM workloads WHERE id = ?
`, id)

	var w Workload
	if err := row.Scan(
		&w.ID, &w.ClusterID, &w.NodeID, &w.Name, &w.Kind, &w.Status,
		&w.CreatedAt, &w.CreateUserID, &w.UpdatedAt, &w.UpdateUserID,
	); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WorkloadRepository) ListByCluster(ctx context.Context, clusterID string) ([]Workload, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, cluster_id, node_id, name, kind, status,
created_at, create_user_id, updated_at, update_user_id
FROM workloads WHERE cluster_id = ?
`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Workload
	for rows.Next() {
		var w Workload
		if err := rows.Scan(
			&w.ID, &w.ClusterID, &w.NodeID, &w.Name, &w.Kind, &w.Status,
			&w.CreatedAt, &w.CreateUserID, &w.UpdatedAt, &w.UpdateUserID,
		); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, nil
}

func (r *WorkloadRepository) ListByNode(ctx context.Context, nodeID string) ([]Workload, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, cluster_id, node_id, name, kind, status,
created_at, create_user_id, updated_at, update_user_id
FROM workloads WHERE node_id = ?
`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Workload
	for rows.Next() {
		var w Workload
		if err := rows.Scan(
			&w.ID, &w.ClusterID, &w.NodeID, &w.Name, &w.Kind, &w.Status,
			&w.CreatedAt, &w.CreateUserID, &w.UpdatedAt, &w.UpdateUserID,
		); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, nil
}
