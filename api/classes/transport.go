package classes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/Vinubaba/SANTC-API/api/store"

	"github.com/Vinubaba/SANTC-API/api/ageranges"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type ClassTransport struct {
	Id          string                      `json:"id"`
	DaycareId   string                      `json:"daycareId"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	ImageUri    string                      `json:"imageUri"`
	AgeRange    ageranges.AgeRangeTransport `json:"ageRange"`
}

type HandlerFactory struct {
	Service Service `inject:""`
}

func (h *HandlerFactory) Add(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeAddEndpoint(h.Service),
		decodeClassTransport,
		shared.EncodeResponse201,
		opts...,
	)
}

func (h *HandlerFactory) Get(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service),
		decodeGetOrDeleteClassTransport,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) Delete(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service),
		decodeGetOrDeleteClassTransport,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) Update(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service),
		decodeUpdateClassRequest,
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
		req := request.(ClassTransport)
		class, err := svc.AddClass(ctx, req)
		if err != nil {
			return nil, err
		}
		return storeToTransport(class), nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClassTransport)
		class, err := svc.GetClass(ctx, req)
		if err != nil {
			return nil, err
		}
		return storeToTransport(class), nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClassTransport)
		if err := svc.DeleteClass(ctx, req); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		classes, err := svc.ListClasses(ctx)
		if err != nil {
			return nil, err
		}
		classesRet := []ClassTransport{}

		for _, class := range classes {
			classesRet = append(classesRet, storeToTransport(class))
		}

		return classesRet, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClassTransport)

		class, err := svc.UpdateClass(ctx, req)
		if err != nil {
			return nil, err
		}

		return storeToTransport(class), nil
	}
}

func decodeClassTransport(_ context.Context, r *http.Request) (interface{}, error) {
	var request ClassTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetOrDeleteClassTransport(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	classId, ok := vars["classId"]
	if !ok {
		return nil, ErrBadRouting
	}
	return ClassTransport{Id: classId}, nil
}

func decodeUpdateClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["classId"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request ClassTransport
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
	case store.ErrClassNotFound:
		w.WriteHeader(http.StatusNotFound)
	case store.ErrAgeRangeNotFound, ErrEmptyAgeRange, store.ErrClassNameAlreadyExists:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func transportToStore(request ClassTransport) store.Class {
	return store.Class{
		DaycareId:   store.DbNullString(request.DaycareId),
		ImageUri:    store.DbNullString(request.ImageUri),
		ClassId:     store.DbNullString(request.Id),
		Description: store.DbNullString(request.Description),
		Name:        store.DbNullString(request.Name),
		AgeRangeId:  store.DbNullString(request.AgeRange.Id),
		AgeRange: store.AgeRange{
			AgeRangeId: store.DbNullString(request.AgeRange.Id),
			DaycareId:  store.DbNullString(request.AgeRange.DaycareId),
			Stage:      store.DbNullString(request.AgeRange.Stage),
			Min:        request.AgeRange.Min,
			MinUnit:    store.DbNullString(request.AgeRange.MinUnit),
			Max:        request.AgeRange.Max,
			MaxUnit:    store.DbNullString(request.AgeRange.MaxUnit),
		},
	}
}

func storeToTransport(class store.Class) ClassTransport {
	return ClassTransport{
		Id:          class.ClassId.String,
		DaycareId:   class.DaycareId.String,
		ImageUri:    class.ImageUri.String,
		Name:        class.Name.String,
		Description: class.Description.String,
		AgeRange: ageranges.AgeRangeTransport{
			Id:        class.AgeRange.AgeRangeId.String,
			DaycareId: class.AgeRange.DaycareId.String,
			Stage:     class.AgeRange.Stage.String,
			Min:       class.AgeRange.Min,
			MinUnit:   class.AgeRange.MinUnit.String,
			Max:       class.AgeRange.Max,
			MaxUnit:   class.AgeRange.MaxUnit.String,
		},
	}
}
