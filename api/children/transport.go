package children

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/Vinubaba/SANTC-API/api/store"

	"github.com/araddon/dateparse"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type ChildTransport struct {
	Id                  string                        `json:"id"`
	DaycareId           string                        `json:"daycareId"`
	ClassId             string                        `json:"classId"`
	FirstName           string                        `json:"firstName"`
	LastName            string                        `json:"lastName"`
	BirthDate           string                        `json:"birthDate"` // dd/mm/yyyy
	Gender              string                        `json:"gender"`
	ImageUri            string                        `json:"imageUri"`
	StartDate           string                        `json:"startDate"` // dd/mm/yyyy
	Notes               string                        `json:"notes"`
	Allergies           []AllergyTransport            `json:"allergies"`
	ResponsibleId       string                        `json:"responsibleId"`
	Relationship        string                        `json:"relationship"`
	SpecialInstructions []SpecialInstructionTransport `json:"specialInstructions"`
}

type AllergyTransport struct {
	Id          string `json:"id"`
	Allergy     string `json:"allergy"`
	Instruction string `json:"instruction"`
}

type SpecialInstructionTransport struct {
	Id          string `json:"id"`
	ChildId     string `json:"childId"`
	Instruction string `json:"instruction"`
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

func ignorePayload(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case ErrNoParent, store.ErrSetResponsible, ErrUpdateDaycare, store.ErrClassNotFound:
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
	ret := ChildTransport{
		Id:            child.ChildId.String,
		DaycareId:     child.DaycareId.String,
		ClassId: child.ClassId.String,
		LastName:      child.LastName.String,
		FirstName:     child.FirstName.String,
		BirthDate:     child.BirthDate.UTC().String(),
		Gender:        child.Gender.String,
		ImageUri:      child.ImageUri.String,
		StartDate:     child.StartDate.UTC().String(),
		Notes:         child.Notes.String,
		ResponsibleId: child.ResponsibleId.String,
		Relationship:  child.Relationship.String,
	}

	for _, allergy := range child.Allergies {
		ret.Allergies = append(ret.Allergies, AllergyTransport{
			Id:          allergy.AllergyId.String,
			Allergy:     allergy.Allergy.String,
			Instruction: allergy.Instruction.String,
		})
	}

	for _, instruction := range child.SpecialInstructions {
		ret.SpecialInstructions = append(ret.SpecialInstructions, SpecialInstructionTransport{
			Id:          instruction.SpecialInstructionId.String,
			ChildId:     instruction.ChildId.String,
			Instruction: instruction.Instruction.String,
		})
	}
	return ret
}

func transportToStore(request ChildTransport, strict bool) (store.Child, error) {
	var birthDate, startDate time.Time
	var err error

	// In case of AddChild, dates are mandatory while in case of update they are not
	if strict || (!strict && request.BirthDate != "") {
		birthDate, err = dateparse.ParseIn(request.BirthDate, time.UTC)
		if err != nil {
			return store.Child{}, err
		}
	}

	if strict || (!strict && request.StartDate != "") {
		startDate, err = dateparse.ParseIn(request.StartDate, time.UTC)
		if err != nil {
			return store.Child{}, err
		}
	}

	child := store.Child{
		ChildId:       store.DbNullString(request.Id),
		DaycareId:     store.DbNullString(request.DaycareId),
		ClassId: store.DbNullString(request.ClassId),
		BirthDate:     birthDate,
		FirstName:     store.DbNullString(request.FirstName),
		LastName:      store.DbNullString(request.LastName),
		Gender:        store.DbNullString(request.Gender),
		ImageUri:      store.DbNullString(request.ImageUri),
		Notes:         store.DbNullString(request.Notes),
		StartDate:     startDate,
		ResponsibleId: store.DbNullString(request.ResponsibleId),
		Relationship:  store.DbNullString(request.Relationship),
	}
	for _, specialInstruction := range request.SpecialInstructions {
		instructionToCreate := store.SpecialInstruction{Instruction: store.DbNullString(specialInstruction.Instruction)}
		child.SpecialInstructions = append(child.SpecialInstructions, instructionToCreate)
	}
	for _, allergy := range request.Allergies {
		allergyToCreate := store.Allergy{Allergy: store.DbNullString(allergy.Allergy), Instruction: store.DbNullString(allergy.Instruction)}
		child.Allergies = append(child.Allergies, allergyToCreate)
	}
	return child, nil
}
