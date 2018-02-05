package children

import (
	"context"
	"encoding/json"
	"net/http"

	auth "github.com/DigitalFrameworksLLC/teddycare/authentication"
	"github.com/DigitalFrameworksLLC/teddycare/shared"
	"github.com/DigitalFrameworksLLC/teddycare/store"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
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
	Image         string   `json:"image"`
	Allergies     []string `json:"allergies"`
	ResponsibleId string   `json:"responsibleId,omitempty"`
	Relationship  string   `json:"relationship,omitempty"`
}

// MakeHandler returns a handler for the booking service.
func MakeHandler(r *mux.Router, svc Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	addChildHandler := kithttp.NewServer(
		makeAddEndpoint(svc),
		decodeChildTransport,
		shared.EncodeResponse201,
		opts...,
	)

	deleteChildHandler := kithttp.NewServer(
		makeDeleteEndpoint(svc),
		decodeGetOrDeleteChildTransport,
		shared.EncodeResponse204,
		opts...,
	)

	getChildHandler := kithttp.NewServer(
		makeGetEndpoint(svc),
		decodeGetOrDeleteChildTransport,
		shared.EncodeResponse200,
		opts...,
	)

	updateChildHandler := kithttp.NewServer(
		makeUpdateEndpoint(svc),
		decodeUpdateChildRequest,
		shared.EncodeResponse200,
		opts...,
	)

	listChildHandler := kithttp.NewServer(
		makeListEndpoint(svc),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)

	r.Handle("/children", auth.Roles(addChildHandler, store.ROLE_OFFICE_MANAGER, store.ROLE_ADULT, store.ROLE_ADMIN)).Methods("POST")
	r.Handle("/children", auth.Roles(listChildHandler, store.ROLE_OFFICE_MANAGER, store.ROLE_ADULT, store.ROLE_ADMIN)).Methods(http.MethodGet)
	r.Handle("/children/{childId}", updateChildHandler).Methods(http.MethodPatch)
	r.Handle("/children/{childId}", deleteChildHandler).Methods(http.MethodDelete)
	r.Handle("/children/{childId}", getChildHandler).Methods(http.MethodGet)
	return r
}

func makeAddEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ChildTransport)
		child, err := svc.AddChild(ctx, req)
		if err != nil {
			return shared.NewError(err.Error()), nil
		}

		ret := ChildTransport{
			Id:            child.ChildId,
			LastName:      child.LastName,
			FirstName:     child.FirstName,
			BirthDate:     child.BirthDate.UTC().String(),
			Gender:        child.Gender,
			Image:         child.ImageUri,
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
			return shared.NewError(err.Error()), nil
		}

		currentChild := ChildTransport{
			Id:        child.ChildId,
			Image:     child.ImageUri,
			Gender:    child.Gender,
			BirthDate: child.BirthDate.UTC().String(),
			FirstName: child.FirstName,
			LastName:  child.LastName,
		}
		allergies, err := svc.FindAllergiesOfChild(ctx, currentChild.Id)
		if err != nil {
			return shared.NewError(err.Error()), nil
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
				Image:     child.ImageUri,
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
			Image:     child.ImageUri,
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
