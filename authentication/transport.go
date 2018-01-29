package authentication

import (
	"arthurgustin.fr/teddycare/shared"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
)

type AuthenticateTransport struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

// MakeHandler returns a handler for the booking service.
func MakeAuthHandler(r *mux.Router, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	authenticateHandler := kithttp.NewServer(
		makeEndpoint(svc),
		decodeRequest,
		shared.EncodeResponse200,
		opts...,
	)

	r.Handle("/authenticate", authenticateHandler).Methods("POST")
	return r
}

func makeEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AuthenticateTransport)

		token, err := svc.Authenticate(ctx, req)
		if err != nil {
			return nil, err
		}

		return token, nil
	}
}

func decodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request AuthenticateTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
