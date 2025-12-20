package database

import (
	"context"
	"database/sql"
	"time"
)

type Cluster struct {
	ID           string
	Name         string
	State        string
	CreatedAt    time.Time
	CreateUserID *string
	UpdatedAt    time.Time
	UpdateUserID *string
}

type ClusterRepository struct {
	exec sqlExecutor
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func NewClusterRepository(db *sql.DB) *ClusterRepository {
	return &ClusterRepository{exec: db}
}

func NewClusterRepositoryTx(tx *sql.Tx) *ClusterRepository {
	return &ClusterRepository{exec: tx}
}

func (r *ClusterRepository) Create(ctx context.Context, c *Cluster) error {
	_, err := r.exec.ExecContext(ctx, `INSERT INTO clusters (id, name, state, create_user_id)
	VALUES (?, ?, ?, ?)`, c.ID, c.Name, c.State, c.CreateUserID)
	return err
}

func (r *ClusterRepository) UpdateByID(ctx context.Context, c *Cluster) error {
	_, err := r.exec.ExecContext(ctx, `UPDATE clusters
	SET name = ?, state = ?, updated_at = CURRENT_TIMESTAMP, update_user_id = ?
	WHERE id = ?`, c.Name, c.State, c.UpdateUserID, c.ID)
	return err
}

func (r *ClusterRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM clusters WHERE id = ?`, id)
	return err
}

func (r *ClusterRepository) GetByID(ctx context.Context, id string) (*Cluster, error) {
	row := r.exec.QueryRowContext(ctx, `SELECT id, name, state, created_at, create_user_id, updated_at, update_user_id
	FROM clusters WHERE id = ?`, id)

	var c Cluster
	if err := row.Scan(
		&c.ID, &c.Name, &c.State,
		&c.CreatedAt, &c.CreateUserID,
		&c.UpdatedAt, &c.UpdateUserID,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClusterRepository) GetByName(ctx context.Context, name string) (*Cluster, error) {
	row := r.exec.QueryRowContext(ctx, `SELECT id, name, state, created_at, create_user_id, updated_at, update_user_id
	FROM clusters WHERE name = ?`, name)

	var c Cluster
	if err := row.Scan(
		&c.ID, &c.Name, &c.State,
		&c.CreatedAt, &c.CreateUserID,
		&c.UpdatedAt, &c.UpdateUserID,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClusterRepository) Count(ctx context.Context) (int, error) {
	row := r.exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM clusters`)
	var n int
	return n, row.Scan(&n)
}
