package cluster

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) InitCluster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// result, err := h.service.InitCluster(r.Context(), &req)
	// if err != nil {
	// 	http.Error(w, err.Error(), 409)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{Success: true})

}
