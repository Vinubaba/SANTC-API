package childs

import (
	"arthurgustin.fr/teddycare/shared"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"net/http"

	"arthurgustin.fr/teddycare/api"
	"arthurgustin.fr/teddycare/store"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
)

type ChildRequest struct {
	Id            string `json:"id"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	BirthDate     string `json:"birthDate"` // dd/mm/yyyy
	ResponsibleId string `json:"responsibleId"`
	Relationship  string `json:"relationship"`
}

type ChildResponse struct {
	Id        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	BirthDate string `json:"birthDate"` // dd/mm/yyyy
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
		shared.EncodeResponse201,
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
		child, err := svc.AddChild(ctx, req)
		if err != nil {
			return shared.NewError(err.Error()), nil
		}

		return ChildResponse{
			Id:        child.ChildId,
			LastName:  child.LastName,
			FirstName: child.FirstName,
			BirthDate: child.BirthDate.UTC().String(),
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

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case ErrNoParent:
		w.WriteHeader(http.StatusBadRequest)
	case store.ErrUserNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
