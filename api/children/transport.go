package children

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Vinubaba/SANTC-API/api/shared"
	. "github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

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

func (h *HandlerFactory) AddPhoto(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeAddPhotoEndpoint(h.Service),
		decodePhotoRequest,
		shared.EncodeResponse201,
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

		return storeToTransport(child), nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ChildTransport)
		child, err := svc.GetChild(ctx, req)
		if err != nil {
			return nil, err
		}
		return storeToTransport(child), nil
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
		children, err := svc.ListChildren(ctx)
		if err != nil {
			return nil, err
		}
		childrenRet := []ChildTransport{}

		for _, child := range children {
			childrenRet = append(childrenRet, storeToTransport(child))
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
		return storeToTransport(child), nil
	}
}

func makeAddPhotoEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PhotoRequestTransport)
		if err := svc.AddPhoto(ctx, req); err != nil {
			return nil, err
		}

		return nil, nil
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
	return ChildTransport{Id: &childId}, nil
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
	request.Id = &id
	return request, nil
}

func decodePhotoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	childId, ok := vars["childId"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request PhotoRequestTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	request.ChildId = &childId
	return request, nil
}

func ignorePayload(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case ErrNoParent, store.ErrSetResponsible, ErrUpdateDaycare, store.ErrClassNotFound, ErrDifferentDaycare:
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

func storeToTransport(child store.Child) ChildTransport {
	birthDate := child.BirthDate.UTC().String()
	startDate := child.StartDate.UTC().String()
	ret := ChildTransport{
		Id:            &child.ChildId.String,
		DaycareId:     &child.DaycareId.String,
		ClassId:       &child.ClassId.String,
		LastName:      &child.LastName.String,
		FirstName:     &child.FirstName.String,
		BirthDate:     &birthDate,
		Gender:        &child.Gender.String,
		ImageUri:      &child.ImageUri.String,
		StartDate:     &startDate,
		Notes:         &child.Notes.String,
		ResponsibleId: &child.ResponsibleId.String,
		Relationship:  &child.Relationship.String,
	}

	for _, allergy := range child.Allergies {
		ret.Allergies = append(ret.Allergies, AllergyTransport{
			Id:          &allergy.AllergyId.String,
			Allergy:     &allergy.Allergy.String,
			Instruction: &allergy.Instruction.String,
		})
	}

	for _, instruction := range child.SpecialInstructions {
		ret.SpecialInstructions = append(ret.SpecialInstructions, SpecialInstructionTransport{
			Id:          &instruction.SpecialInstructionId.String,
			ChildId:     &instruction.ChildId.String,
			Instruction: &instruction.Instruction.String,
		})
	}
	ret.Schedule.Id = &child.Schedule.ScheduleId.String
	ret.Schedule.MondayStart = &child.Schedule.MondayStart.String
	ret.Schedule.MondayEnd = &child.Schedule.MondayEnd.String
	ret.Schedule.TuesdayStart = &child.Schedule.TuesdayStart.String
	ret.Schedule.TuesdayEnd = &child.Schedule.TuesdayEnd.String
	ret.Schedule.WednesdayStart = &child.Schedule.WednesdayStart.String
	ret.Schedule.WednesdayEnd = &child.Schedule.WednesdayEnd.String
	ret.Schedule.ThursdayStart = &child.Schedule.ThursdayStart.String
	ret.Schedule.ThursdayEnd = &child.Schedule.ThursdayEnd.String
	ret.Schedule.FridayStart = &child.Schedule.FridayStart.String
	ret.Schedule.FridayEnd = &child.Schedule.FridayEnd.String
	ret.Schedule.SaturdayStart = &child.Schedule.SaturdayStart.String
	ret.Schedule.SaturdayEnd = &child.Schedule.SaturdayEnd.String
	ret.Schedule.SundayStart = &child.Schedule.SundayStart.String
	ret.Schedule.SundayEnd = &child.Schedule.SundayEnd.String
	ret.Schedule.WalkIn = &child.Schedule.WalkIn.Bool
	return ret
}
