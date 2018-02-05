package children

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DigitalFrameworksLLC/teddycare/shared"
	"github.com/DigitalFrameworksLLC/teddycare/store"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type ChildTransport struct {
	Id            string   `json:"id"`
	FirstName     string   `json:"firstName"`
	LastName      string   `json:"lastName"`
	BirthDate     string   `json:"birthDate"` // dd/mm/yyyy
	Gender        string   `json:"gender"`
	ImageUri      string   `json:"imageUri"`
	Allergies     []string `json:"allergies"`
	ResponsibleId string   `json:"responsibleId,omitempty"`
	Relationship  string   `json:"relationship,omitempty"`
}

type HandlerFactory struct {
	Service Service `inject:""`
}

func (h *HandlerFactory) Add(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeAddEndpoint(h.Service),
		decodeChildTransport,
		shared.EncodeResponse201,
		opts...,
	)
}

func (h *HandlerFactory) Get(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service),
		decodeGetOrDeleteChildTransport,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) Delete(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service),
		decodeGetOrDeleteChildTransport,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) Update(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service),
		decodeUpdateChildRequest,
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
		req := request.(ChildTransport)
		child, err := svc.AddChild(ctx, req)
		if err != nil {
			return nil, err
		}

		ret := ChildTransport{
			Id:            child.ChildId,
			LastName:      child.LastName,
			FirstName:     child.FirstName,
			BirthDate:     child.BirthDate.UTC().String(),
			Gender:        child.Gender,
			ImageUri:      child.ImageUri,
			ResponsibleId: req.ResponsibleId,
			Allergies:     req.Allergies,
		}
		return ret, nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ChildTransport)
		child, err := svc.GetChild(ctx, req)
		if err != nil {
			return nil, err
		}

		currentChild := ChildTransport{
			Id:        child.ChildId,
			ImageUri:  child.ImageUri,
			Gender:    child.Gender,
			BirthDate: child.BirthDate.UTC().String(),
			FirstName: child.FirstName,
			LastName:  child.LastName,
		}
		allergies, err := svc.FindAllergiesOfChild(ctx, currentChild.Id)
		if err != nil {
			return nil, err
		}
		for _, allergy := range allergies {
			currentChild.Allergies = append(currentChild.Allergies, allergy.Allergy)
		}

		return currentChild, nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ChildTransport)
		if err := svc.DeleteChild(ctx, req); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		children, err := svc.ListChild(ctx)
		if err != nil {
			return nil, err
		}
		childrenRet := []ChildTransport{}

		for _, child := range children {
			currentChild := ChildTransport{
				Id:        child.ChildId,
				ImageUri:  child.ImageUri,
				Gender:    child.Gender,
				BirthDate: child.BirthDate.UTC().String(),
				FirstName: child.FirstName,
				LastName:  child.LastName,
			}
			allergies, err := svc.FindAllergiesOfChild(ctx, child.ChildId)
			if err != nil {
				return nil, err
			}
			for _, allergy := range allergies {
				currentChild.Allergies = append(currentChild.Allergies, allergy.Allergy)
			}
			childrenRet = append(childrenRet, currentChild)
		}

		return childrenRet, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ChildTransport)

		child, err := svc.UpdateChild(ctx, req)
		if err != nil {
			return nil, err
		}

		ret := ChildTransport{
			Id:        child.ChildId,
			FirstName: child.FirstName,
			LastName:  child.LastName,
			Gender:    child.Gender,
			BirthDate: child.BirthDate.UTC().String(),
			ImageUri:  child.ImageUri,
		}

		allergies, err := svc.FindAllergiesOfChild(ctx, child.ChildId)
		if err != nil {
			return nil, err
		}
		for _, allergy := range allergies {
			ret.Allergies = append(ret.Allergies, allergy.Allergy)
		}

		return ret, nil
	}
}

func decodeChildTransport(_ context.Context, r *http.Request) (interface{}, error) {
	var request ChildTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetOrDeleteChildTransport(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	childId, ok := vars["childId"]
	if !ok {
		return nil, ErrBadRouting
	}
	return ChildTransport{Id: childId}, nil
}

func decodeUpdateChildRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["childId"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request ChildTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	request.Id = id
	return request, nil
}

type User struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func ignorePayload(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case ErrNoParent, ErrSetResponsible:
		w.WriteHeader(http.StatusBadRequest)
	case store.ErrUserNotFound, store.ErrChildNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
