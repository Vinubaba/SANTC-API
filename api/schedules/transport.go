package schedules

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
		decodeScheduleTransport,
		shared.EncodeResponse201,
		opts...,
	)
}

func (h *HandlerFactory) Get(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service),
		decodeGetOrDeleteScheduleTransport,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) Delete(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service),
		decodeGetOrDeleteScheduleTransport,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) Update(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service),
		decodeUpdateSchedulesRequest,
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
		req := request.(ScheduleTransport)
		schedule, err := svc.AddSchedule(ctx, req)
		if err != nil {
			return nil, err
		}
		return dbSchedulesToTransportSchedules(schedule), nil
	}
}

func makeGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ScheduleTransport)
		schedule, err := svc.GetSchedule(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbSchedulesToTransportSchedules(schedule), nil
	}
}

func makeDeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ScheduleTransport)
		if err := svc.DeleteSchedule(ctx, req); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func makeListEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		schedules, err := svc.ListSchedules(ctx)
		if err != nil {
			return nil, err
		}
		schedulesRet := []ScheduleTransport{}

		for _, schedule := range schedules {
			schedulesRet = append(schedulesRet, dbSchedulesToTransportSchedules(schedule))
		}

		return schedulesRet, nil
	}
}

func makeUpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ScheduleTransport)
		schedule, err := svc.UpdateSchedule(ctx, req)
		if err != nil {
			return nil, err
		}

		return dbSchedulesToTransportSchedules(schedule), nil
	}
}

func dbSchedulesToTransportSchedules(schedule store.Schedule) ScheduleTransport {
	return ScheduleTransport{
		Id:             &schedule.ScheduleId.String,
		WalkIn:         &schedule.WalkIn.Bool,
		MondayStart:    &schedule.MondayStart.String,
		MondayEnd:      &schedule.MondayEnd.String,
		TuesdayStart:   &schedule.TuesdayStart.String,
		TuesdayEnd:     &schedule.TuesdayEnd.String,
		WednesdayStart: &schedule.WednesdayStart.String,
		WednesdayEnd:   &schedule.WednesdayEnd.String,
		ThursdayStart:  &schedule.ThursdayStart.String,
		ThursdayEnd:    &schedule.ThursdayEnd.String,
		FridayStart:    &schedule.FridayStart.String,
		FridayEnd:      &schedule.FridayEnd.String,
		SaturdayStart:  &schedule.SaturdayStart.String,
		SaturdayEnd:    &schedule.SaturdayEnd.String,
		SundayStart:    &schedule.SundayStart.String,
		SundayEnd:      &schedule.SundayEnd.String,
	}
}

func decodeScheduleTransport(_ context.Context, r *http.Request) (interface{}, error) {
	var request ScheduleTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	vars := mux.Vars(r)
	teacherId, teacherFound := vars["teacherId"]
	if teacherFound {
		request.TeacherId = &teacherId
	}
	childId, childFound := vars["childId"]
	if childFound {
		request.ChildId = &childId
	}

	if !teacherFound && !childFound {
		return nil, ErrBadRouting
	}

	return request, nil
}

func decodeGetOrDeleteScheduleTransport(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	teacherId, teacherFound := vars["teacherId"]
	childId, childFound := vars["childId"]
	if !teacherFound && !childFound {
		return nil, ErrBadRouting
	}
	scheduleId, ok := vars["scheduleId"]
	if !ok {
		return nil, ErrBadRouting
	}

	return ScheduleTransport{
		Id:        &scheduleId,
		TeacherId: &teacherId,
		ChildId:   &childId}, nil
}

func decodeUpdateSchedulesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request ScheduleTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	vars := mux.Vars(r)
	teacherId, teacherFound := vars["teacherId"]
	if teacherFound {
		request.TeacherId = &teacherId
	}
	childId, childFound := vars["childId"]
	if childFound {
		request.ChildId = &childId
	}

	id, scheduleIdFound := vars["scheduleId"]
	if scheduleIdFound {
		request.Id = &id
	}

	if (!teacherFound && !childFound) || !scheduleIdFound {
		return nil, ErrBadRouting
	}

	return request, nil
}

func ignorePayload(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch errors.Cause(err) {
	case store.ErrScheduleNotFound, store.ErrChildNotFound, store.ErrUserNotFound:
		w.WriteHeader(http.StatusNotFound)
	case ErrEmptySchedule, ErrBadTimeFormat:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
