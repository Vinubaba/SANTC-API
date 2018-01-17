package api

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
)

type ChildHandler struct {
	Manager interface {
		Add(ctx context.Context) error
	} `inject:""`
}

func (h *ChildHandler) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.Manager.Add(ctx)

	if err != nil {
		switch errors.Cause(err) {
		default:
			writeJSON(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, "created", http.StatusCreated)
}
