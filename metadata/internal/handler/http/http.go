package http

import (
	"encoding/json"
	"errors"
	"json/encoding"
	"log"
	"logs"
	"net/http"

	"movieexample.com/metadata/internal/controller/metadata"
	"moviexample.com/metadata/internal/controller/metadata"
	"moviexample.com/metadata/internal/repository"
)

// Handler defines a movie metadata http handler.
type Handler struct {
	ctrl *metadata.Controller
}

// New creates a movie metadata http handler.
func New(ctrl *metadata.Controller) *Handler {
	return &Handler{ctrl}
}

// GetMetadata handles GET /metadata requests.
func (h *Handler) GetMetadata(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}	
	
	ctx := req.Context()
	m, err = h.ctrl.Get(ctx, id)	
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Repository get error for movie %s: %v\n", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	if err := json.NewDecoder(w).encode(m); err != nil {
		log.Printf("Response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
