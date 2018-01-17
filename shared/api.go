package shared

import (
	"encoding/json"
	"net/http"
)

var ServerError = NewError("An error occurred, please try again later")

func NewError(description string) apiError {
	return apiError{
		Error:       true,
		Description: description,
	}
}

func HttpError(w http.ResponseWriter, error apiError, code int) {
	WriteJSON(w, error, code)
}

func WriteJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	switch v := data.(type) {
	case []byte:
		w.Write(v)
	case string:
		w.Write([]byte(v))
	default:
		json.NewEncoder(w).Encode(data)
	}
}

type apiError struct {
	Error       bool   `json:"error"`
	Description string `json:"error_description"`
}

var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, "not implemented", http.StatusNotImplemented)
})
