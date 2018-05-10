package daycares

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Vinubaba/SANTC-API/shared"
	"github.com/Vinubaba/SANTC-API/store"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting     = errors.New("inconsistent mapping between route and handler (programmer error)")
	ErrNotImplemented = errors.New("not implemented yet")
)

type DaycareTransport struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Address_1 string `json:"address_1"`
	Address_2 string `json:"address_2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
}

type HandlerFactory struct {
	Service Service `inject:""`
}

func (h *HandlerFactory) Add(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeAddEndpoint(h.Service),
		decodeDaycareTransport,
		shared.EncodeResponse201,
		opts...,
	)
}

func (h *HandlerFactory) Get(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service),
		decodeGetOrDeleteDaycareTransport,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) Delete(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service),
		decodeGetOrDeleteDaycareTransport,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) Update(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service),
		decodeUpdateDaycareRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) List(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeListEndpoint(h.Service),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)
}

func makeAddEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DaycareTransport)
		daycare, err := svc.AddDaycare(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbDaycareToTransportDaycare(daycare), nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DaycareTransport)
		daycare, err := svc.GetDaycare(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbDaycareToTransportDaycare(daycare), nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, ErrNotImplemented
		/*req := request.(DaycareTransport)
		if err := svc.DeleteDaycare(ctx, req); err != nil {
			return nil, err
		}
		return nil, nil*/
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		daycares, err := svc.ListDaycare(ctx)
		if err != nil {
			return nil, err
		}
		daycaresRet := []DaycareTransport{}

		for _, daycare := range daycares {
			daycaresRet = append(daycaresRet, dbDaycareToTransportDaycare(daycare))
		}

		return daycaresRet, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DaycareTransport)
		daycare, err := svc.UpdateDaycare(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbDaycareToTransportDaycare(daycare), nil
	}
}

func dbDaycareToTransportDaycare(daycare store.Daycare) DaycareTransport {
	return DaycareTransport{
		Id:        daycare.DaycareId.String,
		Address_1: daycare.Address_1.String,
		Address_2: daycare.Address_2.String,
		City:      daycare.City.String,
		State:     daycare.State.String,
		Zip:       daycare.Zip.String,
		Name:      daycare.Name.String,
	}
}

func decodeDaycareTransport(_ context.Context, r *http.Request) (interface{}, error) {
	var request DaycareTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetOrDeleteDaycareTransport(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["daycareId"]
	if !ok {
		return nil, ErrBadRouting
	}
	return DaycareTransport{Id: id}, nil
}

func decodeUpdateDaycareRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["daycareId"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request DaycareTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	request.Id = id
	return request, nil
}

func ignorePayload(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case store.ErrDaycareNotFound:
		w.WriteHeader(http.StatusNotFound)
	case ErrEmptyDaycare:
		w.WriteHeader(http.StatusBadRequest)
	case ErrNotImplemented:
		w.WriteHeader(http.StatusNotImplemented)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
