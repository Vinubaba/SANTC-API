package adult_responsible

import (
	"arthurgustin.fr/teddycare/store"
	"context"
	"github.com/badoux/checkmail"
	"github.com/pkg/errors"
)

var (
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidPasswordFormat = errors.New("password must be at least 6 characters long")
)

type Service interface {
	AddAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error)
	ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error)
	DeleteAdultResponsible(ctx context.Context, request AdultResponsibleTransport) error
	GetAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error)
	UpdateAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error)
}

type AdultResponsibleService struct {
	Store interface {
		BeginTransaction()
		Commit()
		Rollback()
		AddUser(ctx context.Context, user store.User) (id string, err error)
		AddAdultResponsible(ctx context.Context, adult store.AdultResponsible) (store.AdultResponsible, error)
		ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error)
		DeleteAdultResponsible(ctx context.Context, adultId string) error
		GetAdultResponsible(ctx context.Context, adultId string) (store.AdultResponsible, error)
		UpdateAdultResponsible(ctx context.Context, adult store.AdultResponsible) (store.AdultResponsible, error)
	} `inject:""`
}

func (c AdultResponsibleService) AddAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error) {

	if err := c.validateAddAdultResponsibleRequest(request); err != nil {
		return store.AdultResponsible{}, err
	}

	c.Store.BeginTransaction()

	userId, err := c.Store.AddUser(ctx, store.User{
		Password: request.Password,
		Email:    request.Email,
	})
	if err != nil {
		c.Store.Rollback()
		return store.AdultResponsible{}, errors.New("failed to create user")
	}

	adult, err := c.Store.AddAdultResponsible(ctx, store.AdultResponsible{
		ResponsibleId: userId,
		FirstName:     request.FirstName,
		LastName:      request.LastName,
		Gender:        request.Gender,
		Email:         request.Email,
		Zip:           request.Zip,
		State:         request.State,
		Phone:         request.Phone,
		City:          request.City,
		Addres_1:      request.Addres_1,
		Addres_2:      request.Addres_2,
	})
	if err != nil {
		c.Store.Rollback()
		return store.AdultResponsible{}, errors.Wrap(err, "failed to add adult")
	}
	c.Store.Commit()
	return adult, nil
}

func (c AdultResponsibleService) validateAddAdultResponsibleRequest(req AdultResponsibleTransport) error {
	if err := checkmail.ValidateFormat(req.Email); err != nil {
		return ErrInvalidEmail
	}
	if len(req.Password) < 6 {
		return ErrInvalidPasswordFormat
	}
	return nil
}

func (c AdultResponsibleService) ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error) {
	adults, err := c.Store.ListAdultResponsible(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list adult")
	}

	return adults, nil
}

func (c AdultResponsibleService) DeleteAdultResponsible(ctx context.Context, request AdultResponsibleTransport) error {
	if err := c.Store.DeleteAdultResponsible(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete adult")
	}

	return nil
}

func (c AdultResponsibleService) GetAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error) {
	adult, err := c.Store.GetAdultResponsible(ctx, request.Id)
	if err != nil {
		return adult, errors.Wrap(err, "failed to get adult")
	}

	return adult, nil
}

func (c AdultResponsibleService) UpdateAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error) {
	if request.Email != "" && checkmail.ValidateFormat(request.Email) != nil {
		return store.AdultResponsible{}, ErrInvalidEmail
	}

	adult, err := c.Store.UpdateAdultResponsible(ctx, store.AdultResponsible{
		ResponsibleId: request.Id,
		Email:         request.Email,
		Gender:        request.Gender,
		FirstName:     request.FirstName,
		LastName:      request.LastName,
		Addres_1:      request.Addres_1,
		Addres_2:      request.Addres_2,
		City:          request.City,
		Phone:         request.Phone,
		State:         request.State,
		Zip:           request.Zip,
	})
	if err != nil {
		return adult, errors.Wrap(err, "failed to update adult")
	}

	return adult, nil
}

func NewDefaultService() Service {
	return &AdultResponsibleService{}
}

// ServiceMiddleware is a chainable behavior modifier for adultResponsibleService.
type ServiceMiddleware func(AdultResponsibleService) AdultResponsibleService
