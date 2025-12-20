package cluster

import (
	"database/sql"
	"net/http"
)

func InitModule(mux *http.ServeMux, db *sql.DB) {
	// Initialize services and handlers here
	handler := NewHandler(NewService(db))

	mux.HandleFunc("/cluster/init", handler.InitCluster)
}
