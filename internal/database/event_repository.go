package database

import (
	"context"
	"database/sql"
	"time"
)

type Event struct {
	ID        int64
	ClusterID *string
	NodeID    *string
	Type      string
	Message   string
	CreatedAt time.Time
}

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *Event) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO events (cluster_id, node_id, type, message)
VALUES (?, ?, ?, ?)
`, e.ClusterID, e.NodeID, e.Type, e.Message)
	return err
}

func (r *EventRepository) ListByCluster(ctx context.Context, clusterID string, limit int) ([]Event, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, cluster_id, node_id, type, message, created_at
FROM events WHERE cluster_id = ?
ORDER BY created_at DESC LIMIT ?
`, clusterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID, &e.ClusterID, &e.NodeID,
			&e.Type, &e.Message, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, nil
}
