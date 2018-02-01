package office_manager

import (
	"context"
	"encoding/json"
	"net/http"

	auth "arthurgustin.fr/teddycare/authentication"
	"arthurgustin.fr/teddycare/shared"
	. "arthurgustin.fr/teddycare/store"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type OfficeManagerTransport struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

// MakeHandler returns a handler for the booking service.
func MakeHandler(r *mux.Router, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	addOfficeManagerHandler := kithttp.NewServer(
		makeAddEndpoint(svc),
		decodeAddChildRequest,
		shared.EncodeResponse201,
		opts...,
	)

	listOfficeManagerHandler := kithttp.NewServer(
		makeListEndpoint(svc),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)

	deleteOfficeManagerHandler := kithttp.NewServer(
		makeDeleteEndpoint(svc),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse204,
		opts...,
	)

	getOfficeManagerHandler := kithttp.NewServer(
		makeGetEndpoint(svc),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse200,
		opts...,
	)

	updateOfficeManagerHandler := kithttp.NewServer(
		makeUpdateEndpoint(svc),
		decodeUpdateChildRequest,
		shared.EncodeResponse200,
		opts...,
	)

	r.Handle("/office-managers", auth.Roles(addOfficeManagerHandler, ROLE_ADMIN, ROLE_ADULT)).Methods(http.MethodPost)
	r.Handle("/office-managers", auth.Roles(listOfficeManagerHandler, ROLE_ADMIN)).Methods(http.MethodGet)
	r.Handle("/office-managers/{id}", auth.Roles(getOfficeManagerHandler, ROLE_ADMIN)).Methods(http.MethodGet)
	r.Handle("/office-managers/{id}", auth.Roles(updateOfficeManagerHandler, ROLE_ADMIN)).Methods(http.MethodPatch)
	r.Handle("/office-managers/{id}", auth.Roles(deleteOfficeManagerHandler, ROLE_ADMIN)).Methods(http.MethodDelete)
	return r
}

func makeAddEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OfficeManagerTransport)

		officeManager, err := svc.AddOfficeManager(ctx, req)
		if err != nil {
			return nil, err
		}

		return OfficeManagerTransport{
			Id:       officeManager.OfficeManagerId,
			Email:    officeManager.Email,
			Password: "", // never return the password
		}, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OfficeManagerTransport)

		officeManager, err := svc.UpdateOfficeManager(ctx, req)
		if err != nil {
			return nil, err
		}

		return OfficeManagerTransport{
			Id:       officeManager.OfficeManagerId,
			Email:    officeManager.Email,
			Password: "", // never return the password
		}, nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OfficeManagerTransport)

		if err := svc.DeleteOfficeManager(ctx, req); err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OfficeManagerTransport)

		officeManager, err := svc.GetOfficeManager(ctx, req)
		if err != nil {
			return nil, err
		}

		return OfficeManagerTransport{
			Id:       officeManager.OfficeManagerId,
			Email:    officeManager.Email,
			Password: "", // never return the password
		}, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		officeManagers, err := svc.ListOfficeManager(ctx)
		if err != nil {
			return nil, err
		}

		allAdults := []OfficeManagerTransport{}
		for _, officeManager := range officeManagers {
			allAdults = append(allAdults, OfficeManagerTransport{
				Id:    officeManager.OfficeManagerId,
				Email: officeManager.Email,
			})
		}

		return allAdults, nil
	}
}

func decodeAddChildRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request OfficeManagerTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeUpdateChildRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request OfficeManagerTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	request.Id = id
	return request, nil
}

func decodeGetOrDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return OfficeManagerTransport{Id: id}, nil
}

func ignorePayload(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case ErrInvalidPasswordFormat, ErrInvalidEmail:
		w.WriteHeader(http.StatusBadRequest)
	case ErrUserNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
