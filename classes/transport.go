package classes

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
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type ClassTransport struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageUri    string `json:"imageUri"`
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

		ret := ClassTransport{
			Id:          class.ClassId.String,
			ImageUri:    class.ImageUri.String,
			Description: class.Description.String,
			Name:        class.Name.String,
		}
		return ret, nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClassTransport)
		class, err := svc.GetClass(ctx, req)
		if err != nil {
			return nil, err
		}

		return ClassTransport{
			Id:          class.ClassId.String,
			ImageUri:    class.ImageUri.String,
			Name:        class.Name.String,
			Description: class.Description.String,
		}, nil
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
		classren, err := svc.ListClass(ctx)
		if err != nil {
			return nil, err
		}
		classrenRet := []ClassTransport{}

		for _, class := range classren {
			currentClass := ClassTransport{
				Id:          class.ClassId.String,
				ImageUri:    class.ImageUri.String,
				Description: class.Description.String,
				Name:        class.Name.String,
			}
			classrenRet = append(classrenRet, currentClass)
		}

		return classrenRet, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClassTransport)

		class, err := svc.UpdateClass(ctx, req)
		if err != nil {
			return nil, err
		}

		return ClassTransport{
			Id:          class.ClassId.String,
			ImageUri:    class.ImageUri.String,
			Name:        class.Name.String,
			Description: class.Description.String,
		}, nil
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
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
