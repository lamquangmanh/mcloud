package database

import (
	"context"
	"database/sql"
	"time"
)

type NodeCertificate struct {
	ID           string
	NodeID       string
	CertPEM      string
	IssuedAt     time.Time
	ExpiresAt    time.Time
	CreatedAt    time.Time
	CreateUserID *string
	UpdatedAt    time.Time
	UpdateUserID *string
}

type NodeCertificateRepository struct {
	db *sql.DB
}

func NewNodeCertificateRepository(db *sql.DB) *NodeCertificateRepository {
	return &NodeCertificateRepository{db: db}
}

func (r *NodeCertificateRepository) Create(ctx context.Context, c *NodeCertificate) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO node_certificates (id, node_id, cert_pem, expires_at, create_user_id)
VALUES (?, ?, ?, ?, ?)
`, c.ID, c.NodeID, c.CertPEM, c.ExpiresAt, c.CreateUserID)
	return err
}

func (r *NodeCertificateRepository) GetByNode(ctx context.Context, nodeID string) ([]NodeCertificate, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, node_id, cert_pem, issued_at, expires_at,
created_at, create_user_id, updated_at, update_user_id
FROM node_certificates WHERE node_id = ?
`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []NodeCertificate
	for rows.Next() {
		var c NodeCertificate
		if err := rows.Scan(
			&c.ID, &c.NodeID, &c.CertPEM, &c.IssuedAt, &c.ExpiresAt,
			&c.CreatedAt, &c.CreateUserID, &c.UpdatedAt, &c.UpdateUserID,
		); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, nil
}

func (r *NodeCertificateRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
DELETE FROM node_certificates WHERE expires_at < ?
`, now)
	return err
}
