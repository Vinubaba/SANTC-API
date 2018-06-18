package ageranges

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type AgeRangeTransport struct {
	Id        string `json:"id"`
	DaycareId string `json:"daycareId"`
	Stage     string `json:"stage"`
	Min       int    `json:"min"`
	MinUnit   string `json:"minUnit"`
	Max       int    `json:"max"`
	MaxUnit   string `json:"maxUnit"`
}

type HandlerFactory struct {
	Service Service `inject:""`
}

func (h *HandlerFactory) Add(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeAddEndpoint(h.Service),
		decodeAgeRangeTransport,
		shared.EncodeResponse201,
		opts...,
	)
}

func (h *HandlerFactory) Get(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service),
		decodeGetOrDeleteAgeRangeTransport,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) Delete(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service),
		decodeGetOrDeleteAgeRangeTransport,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) Update(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service),
		decodeUpdateAgeRangeRequest,
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
		req := request.(AgeRangeTransport)
		ageRange, err := svc.AddAgeRange(ctx, req)
		if err != nil {
			return nil, err
		}
		return dbAgeRangeToTransportAgeRange(ageRange), nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AgeRangeTransport)
		ageRange, err := svc.GetAgeRange(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbAgeRangeToTransportAgeRange(ageRange), nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AgeRangeTransport)
		if err := svc.DeleteAgeRange(ctx, req); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ageRanges, err := svc.ListAgeRange(ctx)
		if err != nil {
			return nil, err
		}
		ageRangesRet := []AgeRangeTransport{}

		for _, ageRange := range ageRanges {
			ageRangesRet = append(ageRangesRet, dbAgeRangeToTransportAgeRange(ageRange))
		}

		return ageRangesRet, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AgeRangeTransport)
		ageRange, err := svc.UpdateAgeRange(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbAgeRangeToTransportAgeRange(ageRange), nil
	}
}

func dbAgeRangeToTransportAgeRange(ageRange store.AgeRange) AgeRangeTransport {
	return AgeRangeTransport{
		Id:        ageRange.AgeRangeId.String,
		DaycareId: ageRange.DaycareId.String,
		Stage:     ageRange.Stage.String,
		Min:       ageRange.Min,
		MinUnit:   ageRange.MinUnit.String,
		Max:       ageRange.Max,
		MaxUnit:   ageRange.MaxUnit.String,
	}
}

func decodeAgeRangeTransport(_ context.Context, r *http.Request) (interface{}, error) {
	var request AgeRangeTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetOrDeleteAgeRangeTransport(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	ageRangeId, ok := vars["ageRangeId"]
	if !ok {
		return nil, ErrBadRouting
	}
	return AgeRangeTransport{Id: ageRangeId}, nil
}

func decodeUpdateAgeRangeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["ageRangeId"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request AgeRangeTransport
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
	case store.ErrAgeRangeNotFound:
		w.WriteHeader(http.StatusNotFound)
	case ErrEmptyAgeRange:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
