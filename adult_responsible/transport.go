package adult_responsible

import (
	"arthurgustin.fr/teddycare/shared"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"net/http"

	auth "arthurgustin.fr/teddycare/authentication"
	. "arthurgustin.fr/teddycare/store"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type AdultResponsibleTransport struct {
	Id        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Gender    string `json:"gender"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	Phone     string `json:"phone"`
	Addres_1  string `json:"address_1"`
	Addres_2  string `json:"address_2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
}

// MakeHandler returns a handler for the booking service.
func MakeHandler(r *mux.Router, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	addAdultResponsibleHandler := kithttp.NewServer(
		makeAddEndpoint(svc),
		decodeAddChildRequest,
		shared.EncodeResponse201,
		opts...,
	)

	listAdultResponsibleHandler := kithttp.NewServer(
		makeListEndpoint(svc),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)

	deleteAdultResponsibleHandler := kithttp.NewServer(
		makeDeleteEndpoint(svc),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse204,
		opts...,
	)

	getAdultResponsibleHandler := kithttp.NewServer(
		makeGetEndpoint(svc),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse200,
		opts...,
	)

	updateAdultResponsibleHandler := kithttp.NewServer(
		makeUpdateEndpoint(svc),
		decodeUpdateChildRequest,
		shared.EncodeResponse200,
		opts...,
	)

	r.Handle("/adults", auth.Roles(addAdultResponsibleHandler, ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_TEACHER)).Methods("POST")
	r.Handle("/adults", auth.Roles(listAdultResponsibleHandler, ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_TEACHER)).Methods(http.MethodGet)
	r.Handle("/adults/{id}", auth.Roles(getAdultResponsibleHandler, ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_TEACHER)).Methods(http.MethodGet)
	r.Handle("/adults/{id}", auth.Roles(updateAdultResponsibleHandler, ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_TEACHER)).Methods(http.MethodPatch)
	r.Handle("/adults/{id}", auth.Roles(deleteAdultResponsibleHandler, ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_TEACHER)).Methods(http.MethodDelete)
	return r
}

func makeAddEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AdultResponsibleTransport)

		adult, err := svc.AddAdultResponsible(ctx, req)
		if err != nil {
			return nil, err
		}

		return AdultResponsibleTransport{
			Id:        adult.ResponsibleId,
			FirstName: adult.FirstName,
			LastName:  adult.LastName,
			Email:     adult.Email,
			Gender:    adult.Gender,
			Password:  "", // never return the password
			Addres_1:  adult.Addres_1,
			Addres_2:  adult.Addres_2,
			City:      adult.City,
			Phone:     adult.Phone,
			State:     adult.State,
			Zip:       adult.Zip,
		}, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AdultResponsibleTransport)

		adult, err := svc.UpdateAdultResponsible(ctx, req)
		if err != nil {
			return nil, err
		}

		return AdultResponsibleTransport{
			Id:        adult.ResponsibleId,
			FirstName: adult.FirstName,
			LastName:  adult.LastName,
			Email:     adult.Email,
			Gender:    adult.Gender,
			Password:  "", // never return the password
			Addres_1:  adult.Addres_1,
			Addres_2:  adult.Addres_2,
			City:      adult.City,
			Phone:     adult.Phone,
			State:     adult.State,
			Zip:       adult.Zip,
		}, nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AdultResponsibleTransport)

		if err := svc.DeleteAdultResponsible(ctx, req); err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AdultResponsibleTransport)

		adult, err := svc.GetAdultResponsible(ctx, req)
		if err != nil {
			return nil, err
		}

		return AdultResponsibleTransport{
			Id:        adult.ResponsibleId,
			FirstName: adult.FirstName,
			LastName:  adult.LastName,
			Email:     adult.Email,
			Gender:    adult.Gender,
			Password:  "", // never return the password
			Addres_1:  adult.Addres_1,
			Addres_2:  adult.Addres_2,
			City:      adult.City,
			Phone:     adult.Phone,
			State:     adult.State,
			Zip:       adult.Zip,
		}, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		adults, err := svc.ListAdultResponsible(ctx)
		if err != nil {
			return nil, err
		}

		allAdults := []AdultResponsibleTransport{}
		for _, adult := range adults {
			allAdults = append(allAdults, AdultResponsibleTransport{
				Id:        adult.ResponsibleId,
				Gender:    adult.Gender,
				LastName:  adult.LastName,
				FirstName: adult.FirstName,
				Email:     adult.Email,
			})
		}

		return allAdults, nil
	}
}

func decodeAddChildRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request AdultResponsibleTransport
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
	var request AdultResponsibleTransport
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
	return AdultResponsibleTransport{Id: id}, nil
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
