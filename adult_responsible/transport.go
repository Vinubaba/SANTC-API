package adult_responsible

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

type AdultResponsibleRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Gender    string `json:"gender"`
}

type AdultResponsibleResponse struct {
	AdultResponsibleRequest
	Id string `json:"id"`
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
		encodeResponse,
		opts...,
	)

	listAdultResponsibleHandler := kithttp.NewServer(
		makeListEndpoint(svc),
		ignorePayload,
		encodeResponse,
		opts...,
	)

	r.Handle("/adults", addAdultResponsibleHandler).Methods("POST")
	r.Handle("/adults", listAdultResponsibleHandler).Methods(http.MethodGet)
	r.Handle("/adults/{id}", api.NotImplemented).Methods(http.MethodPatch)
	r.Handle("/adults/{id}", api.NotImplemented).Methods(http.MethodDelete)
	r.Handle("/adults/{id}", api.NotImplemented).Methods(http.MethodGet)
	return r
}

func makeAddEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AdultResponsibleRequest)
		id, err := svc.AddAdultResponsible(ctx, req)
		if err != nil {
			return shared.NewError(err.Error()), nil
		}

		return AdultResponsibleResponse{
			AdultResponsibleRequest: req,
			Id: id,
		}, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		adults, err := svc.ListAdultResponsible(ctx)
		if err != nil {
			return shared.NewError(err.Error()), nil
		}

		allAdults := []AdultResponsibleResponse{}
		for _, adult := range adults {
			allAdults = append(allAdults, AdultResponsibleResponse{
				Id: adult.ResponsibleId,
				AdultResponsibleRequest: AdultResponsibleRequest{
					Gender:    adult.Gender,
					LastName:  adult.LastName,
					FirstName: adult.FirstName,
				},
			})
		}

		return allAdults, nil
	}
}

func decodeAddChildRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request AdultResponsibleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func ignorePayload(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
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
