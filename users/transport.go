package users

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"net/http"

	"arthurgustin.fr/teddycare/store"
	"github.com/DigitalFrameworksLLC/teddycare/shared"

	"fmt"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

type UserTransport struct {
	Id        string   `json:"id"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Gender    string   `json:"gender"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Address_1 string   `json:"address_1"`
	Address_2 string   `json:"address_2"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Zip       string   `json:"zip"`
	ImageUri  string   `json:"imageUri"`
	Roles     []string `json:"roles"`
}

type HandlerFactory struct {
	Service Service `inject:""`
}

func (h *HandlerFactory) Me(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeMeEndpoint(h.Service),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) AddPendingUser(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeAddPendingUserEndpoint(h.Service),
		decodeUserRequest,
		shared.EncodeResponse201,
		opts...,
	)
}

// ADULT

func (h *HandlerFactory) ListAdult(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeListEndpoint(h.Service, store.ROLE_ADULT),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) GetAdult(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service, store.ROLE_ADULT),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) DeleteAdult(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service, store.ROLE_ADULT),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) UpdateAdult(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service, store.ROLE_ADULT),
		decodeUpdateUserRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

// OFFICE MANAGERS

func (h *HandlerFactory) ListOfficeManager(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeListEndpoint(h.Service, store.ROLE_OFFICE_MANAGER),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) GetOfficeManager(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service, store.ROLE_OFFICE_MANAGER),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) DeleteOfficeManager(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service, store.ROLE_OFFICE_MANAGER),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse204,
		opts...,
	)
}

func (h *HandlerFactory) UpdateOfficeManager(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service, store.ROLE_OFFICE_MANAGER),
		decodeUpdateUserRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

// TEACHER

func (h *HandlerFactory) ListTeacher(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeListEndpoint(h.Service, store.ROLE_TEACHER),
		ignorePayload,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) GetTeacher(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeGetEndpoint(h.Service, store.ROLE_TEACHER),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) UpdateTeacher(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeUpdateEndpoint(h.Service, store.ROLE_TEACHER),
		decodeUpdateUserRequest,
		shared.EncodeResponse200,
		opts...,
	)
}

func (h *HandlerFactory) DeleteTeacher(opts []kithttp.ServerOption) *kithttp.Server {
	return kithttp.NewServer(
		makeDeleteEndpoint(h.Service, store.ROLE_TEACHER),
		decodeGetOrDeleteRequest,
		shared.EncodeResponse204,
		opts...,
	)
}

func makeAddEndpoint(svc Service, role string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserTransport)

		createdUser, err := svc.AddUserByRoles(ctx, req, role)
		if err != nil {
			return nil, err
		}

		return UserTransport{
			Id:        createdUser.UserId,
			FirstName: createdUser.FirstName,
			LastName:  createdUser.LastName,
			Email:     createdUser.Email,
			Gender:    createdUser.Gender,
			Address_1: createdUser.Address_1,
			Address_2: createdUser.Address_2,
			City:      createdUser.City,
			Phone:     createdUser.Phone,
			State:     createdUser.State,
			Zip:       createdUser.Zip,
			ImageUri:  createdUser.ImageUri,
		}, nil
	}
}

func makeAddPendingUserEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserTransport)
		return nil, svc.AddPendingUser(ctx, req)
	}
}

func makeUpdateEndpoint(svc Service, role string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserTransport)

		user, err := svc.UpdateUserByRoles(ctx, req, role)
		if err != nil {
			return nil, err
		}

		return UserTransport{
			Id:        user.UserId,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Gender:    user.Gender,
			Address_1: user.Address_1,
			Address_2: user.Address_2,
			City:      user.City,
			Phone:     user.Phone,
			State:     user.State,
			Zip:       user.Zip,
			ImageUri:  user.ImageUri,
			Roles:     user.Roles.ToList(),
		}, nil
	}
}

func makeDeleteEndpoint(svc Service, role string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserTransport)

		if err := svc.DeleteUserByRoles(ctx, req, role); err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeGetEndpoint(svc Service, role string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserTransport)

		user, err := svc.GetUserByRoles(ctx, req, role)
		if err != nil {
			return nil, err
		}

		return UserTransport{
			Id:        user.UserId,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Gender:    user.Gender,
			Address_1: user.Address_1,
			Address_2: user.Address_2,
			City:      user.City,
			Phone:     user.Phone,
			State:     user.State,
			Zip:       user.Zip,
			ImageUri:  user.ImageUri,
			Roles:     user.Roles.ToList(),
		}, nil
	}
}

func makeMeEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		claims := ctx.Value("decoded").(map[string]interface{})
		meId, ok := claims["userId"]
		if !ok {
			return UserTransport{}, errors.New("no id in decoded claim")
		}

		user, err := svc.GetUserByRoles(ctx, UserTransport{
			Id: meId.(string),
		})
		if err != nil {
			return nil, err
		}

		userRoles, err := svc.GetUserRoles(ctx, UserTransport{
			Id: user.UserId,
		})
		roles := make([]string, 0)
		for _, role := range userRoles {
			roles = append(roles, role.Role)
		}

		return UserTransport{
			Id:        user.UserId,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Gender:    user.Gender,
			Address_1: user.Address_1,
			Address_2: user.Address_2,
			City:      user.City,
			Phone:     user.Phone,
			State:     user.State,
			Zip:       user.Zip,
			ImageUri:  user.ImageUri,
			Roles:     roles,
		}, nil
	}
}

func makeListEndpoint(svc Service, roleConstraint string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		users, err := svc.ListUsersByRole(ctx, roleConstraint)
		if err != nil {
			return nil, err
		}

		allUsers := []UserTransport{}
		for _, user := range users {
			allUsers = append(allUsers, UserTransport{
				Id:        user.UserId,
				Gender:    user.Gender,
				LastName:  user.LastName,
				FirstName: user.FirstName,
				Email:     user.Email,
				ImageUri:  user.ImageUri,
				Phone:     user.Phone,
				Zip:       user.Zip,
				State:     user.State,
				City:      user.City,
				Address_1: user.Address_1,
				Address_2: user.Address_2,
				Roles:     user.Roles.ToList(),
			})
		}
		return allUsers, nil
	}
}

func decodeUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request UserTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeUpdateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// get id from url
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	// get informations from payload
	var request UserTransport
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	request.Id = id
	return request, nil
}

func decodeGetOrDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return UserTransport{Id: id}, nil
}

func ignorePayload(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

// encode errors from business-logic
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	fmt.Println(errors.Cause(err).Error())

	switch errors.Cause(err).Error() {
	case ErrInvalidPasswordFormat.Error(), ErrInvalidEmail.Error():
		w.WriteHeader(http.StatusBadRequest)
	case store.ErrUserNotFound.Error():
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
