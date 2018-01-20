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
	AddAdultResponsible(ctx context.Context, request AddAdultResponsibleRequest) (string, error)
	ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error)
	DeleteAdultResponsible(ctx context.Context, request GetOrDeleteAdultResponsibleRequest) error
	GetAdultResponsible(ctx context.Context, request GetOrDeleteAdultResponsibleRequest) (store.AdultResponsible, error)
	UpdateAdultResponsible(ctx context.Context, request UpdateAdultResponsibleRequest) (store.AdultResponsible, error)
}

type AdultResponsibleService struct {
	Store interface {
		BeginTransaction()
		Commit()
		Rollback()
		AddUser(ctx context.Context, user store.User) (id string, err error)
		AddAdultResponsible(ctx context.Context, adult store.AdultResponsible) (id string, err error)
		ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error)
		DeleteAdultResponsible(ctx context.Context, adultId string) error
		GetAdultResponsible(ctx context.Context, adultId string) (store.AdultResponsible, error)
		UpdateAdultResponsible(ctx context.Context, adult store.AdultResponsible) (store.AdultResponsible, error)
	} `inject:""`
}

func (c AdultResponsibleService) AddAdultResponsible(ctx context.Context, request AddAdultResponsibleRequest) (string, error) {

	if err := c.validateAddAdultResponsibleRequest(request); err != nil {
		return "", err
	}

	c.Store.BeginTransaction()

	userId, err := c.Store.AddUser(ctx, store.User{
		Password: request.Password,
		Email:    request.Email,
	})
	if err != nil {
		c.Store.Rollback()
		return "", errors.New("failed to create user")
	}

	responsibleId, err := c.Store.AddAdultResponsible(ctx, store.AdultResponsible{
		ResponsibleId: userId,
		FirstName:     request.FirstName,
		LastName:      request.LastName,
		Gender:        request.Gender,
		Email:         request.Email,
	})
	if err != nil {
		c.Store.Rollback()
		return "", errors.Wrap(err, "failed to add adult")
	}
	c.Store.Commit()
	return responsibleId, nil
}

func (c AdultResponsibleService) validateAddAdultResponsibleRequest(req AddAdultResponsibleRequest) error {
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

func (c AdultResponsibleService) DeleteAdultResponsible(ctx context.Context, request GetOrDeleteAdultResponsibleRequest) error {
	if err := c.Store.DeleteAdultResponsible(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete adult")
	}

	return nil
}

func (c AdultResponsibleService) GetAdultResponsible(ctx context.Context, request GetOrDeleteAdultResponsibleRequest) (store.AdultResponsible, error) {
	adult, err := c.Store.GetAdultResponsible(ctx, request.Id)
	if err != nil {
		return adult, errors.Wrap(err, "failed to get adult")
	}

	return adult, nil
}

func (c AdultResponsibleService) UpdateAdultResponsible(ctx context.Context, request UpdateAdultResponsibleRequest) (store.AdultResponsible, error) {
	if request.Email != "" && checkmail.ValidateFormat(request.Email) != nil {
		return store.AdultResponsible{}, ErrInvalidEmail
	}

	adult, err := c.Store.UpdateAdultResponsible(ctx, store.AdultResponsible{
		ResponsibleId: request.Id,
		Email:         request.Email,
		Gender:        request.Gender,
		FirstName:     request.FirstName,
		LastName:      request.LastName,
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
