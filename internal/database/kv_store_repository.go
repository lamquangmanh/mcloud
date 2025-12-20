package database

import (
	"context"
	"database/sql"
	"time"
)

type KV struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

type KVStoreRepository struct {
	exec sqlExecutor
}

func NewKVStoreRepository(db *sql.DB) *KVStoreRepository {
	return &KVStoreRepository{exec: db}
}

func NewKVStoreRepositoryTx(tx *sql.Tx) *KVStoreRepository {
	return &KVStoreRepository{exec: tx}
}

func (r *KVStoreRepository) Set(ctx context.Context, key, value string) error {
	_, err := r.exec.ExecContext(ctx, `
INSERT INTO kv_store (key, value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
`, key, value)
	return err
}

func (r *KVStoreRepository) Get(ctx context.Context, key string) (*KV, error) {
	row := r.exec.QueryRowContext(ctx, `
SELECT key, value, updated_at FROM kv_store WHERE key = ?
`, key)

	var kv KV
	if err := row.Scan(&kv.Key, &kv.Value, &kv.UpdatedAt); err != nil {
		return nil, err
	}
	return &kv, nil
}

func (r *KVStoreRepository) Delete(ctx context.Context, key string) error {
	_, err := r.exec.ExecContext(ctx, `DELETE FROM kv_store WHERE key = ?`, key)
	return err
}

func (r *KVStoreRepository) List(ctx context.Context) ([]KV, error) {
	rows, err := r.exec.QueryContext(ctx, `
SELECT key, value, updated_at FROM kv_store
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []KV
	for rows.Next() {
		var kv KV
		if err := rows.Scan(&kv.Key, &kv.Value, &kv.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, kv)
	}
	return items, nil
}
