package shared

import (
	"context"
	"encoding/json"
	"net/http"
)

func EncodeResponse200(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)
	if response != nil {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(response)
	}
	return nil
}

func EncodeResponse201(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusCreated)
	if response != nil {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(response)
	}
	return nil
}

func EncodeResponse204(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
