package database

import (
	"context"
	"database/sql"
	"time"
)

type BootstrapToken struct {
	Token        string
	ClusterID    string
	ExpiresAt    time.Time
	Used         bool
	CreatedAt    time.Time
	CreateUserID *string
	UpdatedAt    time.Time
	UpdateUserID *string
}

type BootstrapTokenRepository struct {
	exec sqlExecutor
}

func NewBootstrapTokenRepository(db *sql.DB) *BootstrapTokenRepository {
	return &BootstrapTokenRepository{exec: db}
}

func NewBootstrapTokenRepositoryTx(tx *sql.Tx) *BootstrapTokenRepository {
	return &BootstrapTokenRepository{exec: tx}
}

func (r *BootstrapTokenRepository) Create(ctx context.Context, t *BootstrapToken) error {
	_, err := r.exec.ExecContext(ctx, `
	INSERT INTO bootstrap_tokens (token, cluster_id, expires_at, used, create_user_id)
	VALUES (?, ?, ?, ?, ?)`, t.Token, t.ClusterID, t.ExpiresAt, t.Used, t.CreateUserID)
	return err
}

func (r *BootstrapTokenRepository) MarkUsed(ctx context.Context, token string) error {
	_, err := r.exec.ExecContext(ctx, `UPDATE bootstrap_tokens
	SET used = 1, updated_at = CURRENT_TIMESTAMP
	WHERE token = ?`, token)
	return err
}

func (r *BootstrapTokenRepository) Delete(ctx context.Context, token string) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM bootstrap_tokens WHERE token = ?`, token)
	return err
}

func (r *BootstrapTokenRepository) Get(ctx context.Context, token string) (*BootstrapToken, error) {
	row := r.exec.QueryRowContext(ctx, `SELECT token, cluster_id, expires_at, used,
	created_at, create_user_id, updated_at, update_user_id
	FROM bootstrap_tokens WHERE token = ?
	`, token)

	var t BootstrapToken
	var usedInt int
	if err := row.Scan(
		&t.Token, &t.ClusterID, &t.ExpiresAt, &usedInt,
		&t.CreatedAt, &t.CreateUserID, &t.UpdatedAt, &t.UpdateUserID,
	); err != nil {
		return nil, err
	}
	t.Used = usedInt == 1
	return &t, nil
}

func (r *BootstrapTokenRepository) ListByCluster(ctx context.Context, clusterID string) ([]BootstrapToken, error) {
	rows, err := r.exec.QueryContext(ctx, `
		SELECT token, cluster_id, expires_at, used,
		created_at, create_user_id, updated_at, update_user_id
		FROM bootstrap_tokens WHERE cluster_id = ?
		`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BootstrapToken
	for rows.Next() {
		var t BootstrapToken
		var usedInt int
		if err := rows.Scan(
			&t.Token, &t.ClusterID, &t.ExpiresAt, &usedInt,
			&t.CreatedAt, &t.CreateUserID, &t.UpdatedAt, &t.UpdateUserID,
		); err != nil {
			return nil, err
		}
		t.Used = usedInt == 1
		items = append(items, t)
	}
	return items, nil
}
