package shared

import (
	"context"
	"encoding/json"
	"net/http"
)

func EncodeResponse200(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func EncodeResponse201(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func EncodeResponse204(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusNoContent)
	return json.NewEncoder(w).Encode(response)
}
