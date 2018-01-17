package childs

import (
	"arthurgustin.fr/teddycare/shared"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"net/http"

	"arthurgustin.fr/teddycare/api"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
)

type ChildRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	BirthDate     string `json:"birthDate"` // dd/mm/yyyy
	ResponsibleId string `json:"responsibleId"`
	Relationship  string `json:"relationship"`
}

type ChildResponse struct {
	ChildRequest
	Id string `json:"id"`
}

// MakeHandler returns a handler for the booking service.
func MakeHandler(r *mux.Router, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	addChildHandler := kithttp.NewServer(
		makeAddEndpoint(svc),
		decodeAddChildRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/childs", addChildHandler).Methods("POST")
	r.Handle("/childs", api.NotImplemented).Methods(http.MethodGet)
	r.Handle("/childs/{id}", api.NotImplemented).Methods(http.MethodPatch)
	r.Handle("/childs/{id}", api.NotImplemented).Methods(http.MethodDelete)
	r.Handle("/childs/{id}", api.NotImplemented).Methods(http.MethodGet)
	return r
}

func makeAddEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ChildRequest)
		id, err := svc.AddChild(ctx, req)
		if err != nil {
			return shared.NewError(err.Error()), nil
		}

		return ChildResponse{
			ChildRequest: req,
			Id:           id,
		}, nil
	}
}

func decodeAddChildRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request ChildRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
