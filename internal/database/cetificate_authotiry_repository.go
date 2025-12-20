package database

import (
	"context"
	"database/sql"
	"time"
)

type CertificateAuthority struct {
	ID           string
	ClusterID    string
	CertPEM      string
	KeyPEM       string
	CreatedAt    time.Time
	CreateUserID *string
	UpdatedAt    time.Time
	UpdateUserID *string
}

type CertificateAuthorityRepository struct {
	exec sqlExecutor
}

func NewCertificateAuthorityRepository(db *sql.DB) *CertificateAuthorityRepository {
	return &CertificateAuthorityRepository{exec: db}
}

func NewCertificateAuthorityRepositoryTx(tx *sql.Tx) *CertificateAuthorityRepository {
	return &CertificateAuthorityRepository{exec: tx}
}

func (r *CertificateAuthorityRepository) Create(ctx context.Context, ca *CertificateAuthority) error {
	_, err := r.exec.ExecContext(ctx, `
INSERT INTO certificate_authorities (id, cluster_id, cert_pem, key_pem, create_user_id)
VALUES (?, ?, ?, ?, ?)
`, ca.ID, ca.ClusterID, ca.CertPEM, ca.KeyPEM, ca.CreateUserID)
	return err
}

func (r *CertificateAuthorityRepository) GetByCluster(ctx context.Context, clusterID string) (*CertificateAuthority, error) {
	row := r.exec.QueryRowContext(ctx, `
SELECT id, cluster_id, cert_pem, key_pem,
created_at, create_user_id, updated_at, update_user_id
FROM certificate_authorities WHERE cluster_id = ?
`, clusterID)

	var ca CertificateAuthority
	if err := row.Scan(
		&ca.ID, &ca.ClusterID, &ca.CertPEM, &ca.KeyPEM,
		&ca.CreatedAt, &ca.CreateUserID, &ca.UpdatedAt, &ca.UpdateUserID,
	); err != nil {
		return nil, err
	}
	return &ca, nil
}

func (r *CertificateAuthorityRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM certificate_authorities WHERE id = ?`, id)
	return err
}
