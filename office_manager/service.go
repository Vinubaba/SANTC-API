package office_manager

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
	AddOfficeManager(ctx context.Context, request OfficeManagerTransport) (store.OfficeManager, error)
	ListOfficeManager(ctx context.Context) ([]store.OfficeManager, error)
	DeleteOfficeManager(ctx context.Context, request OfficeManagerTransport) error
	GetOfficeManager(ctx context.Context, request OfficeManagerTransport) (store.OfficeManager, error)
	UpdateOfficeManager(ctx context.Context, request OfficeManagerTransport) (store.OfficeManager, error)
}

type OfficeManagerService struct {
	Store interface {
		BeginTransaction()
		Commit()
		Rollback()
		AddUser(ctx context.Context, user store.User) (id string, err error)
		AddOfficeManager(ctx context.Context, officeManager store.OfficeManager) (store.OfficeManager, error)
		ListOfficeManager(ctx context.Context) ([]store.OfficeManager, error)
		DeleteOfficeManager(ctx context.Context, officeManagerId string) error
		GetOfficeManager(ctx context.Context, officeManagerId string) (store.OfficeManager, error)
		UpdateOfficeManager(ctx context.Context, officeManager store.OfficeManager) (store.OfficeManager, error)
	} `inject:""`
}

func (c OfficeManagerService) AddOfficeManager(ctx context.Context, request OfficeManagerTransport) (store.OfficeManager, error) {

	if err := c.validateAddOfficeManagerRequest(request); err != nil {
		return store.OfficeManager{}, err
	}

	c.Store.BeginTransaction()

	userId, err := c.Store.AddUser(ctx, store.User{
		Password: request.Password,
		Email:    request.Email,
	})
	if err != nil {
		c.Store.Rollback()
		return store.OfficeManager{}, errors.New("failed to create user")
	}

	officeManager, err := c.Store.AddOfficeManager(ctx, store.OfficeManager{
		OfficeManagerId: userId,
		Email:         request.Email,
	})
	if err != nil {
		c.Store.Rollback()
		return store.OfficeManager{}, errors.Wrap(err, "failed to add officeManager")
	}
	c.Store.Commit()
	return officeManager, nil
}

func (c OfficeManagerService) validateAddOfficeManagerRequest(req OfficeManagerTransport) error {
	if err := checkmail.ValidateFormat(req.Email); err != nil {
		return ErrInvalidEmail
	}
	if len(req.Password) < 6 {
		return ErrInvalidPasswordFormat
	}
	return nil
}

func (c OfficeManagerService) ListOfficeManager(ctx context.Context) ([]store.OfficeManager, error) {
	officeManagers, err := c.Store.ListOfficeManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list officeManager")
	}

	return officeManagers, nil
}

func (c OfficeManagerService) DeleteOfficeManager(ctx context.Context, request OfficeManagerTransport) error {
	if err := c.Store.DeleteOfficeManager(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete officeManager")
	}

	return nil
}

func (c OfficeManagerService) GetOfficeManager(ctx context.Context, request OfficeManagerTransport) (store.OfficeManager, error) {
	officeManager, err := c.Store.GetOfficeManager(ctx, request.Id)
	if err != nil {
		return officeManager, errors.Wrap(err, "failed to get officeManager")
	}

	return officeManager, nil
}

func (c OfficeManagerService) UpdateOfficeManager(ctx context.Context, request OfficeManagerTransport) (store.OfficeManager, error) {
	if request.Email != "" && checkmail.ValidateFormat(request.Email) != nil {
		return store.OfficeManager{}, ErrInvalidEmail
	}

	officeManager, err := c.Store.UpdateOfficeManager(ctx, store.OfficeManager{
		OfficeManagerId: request.Id,
		Email:         request.Email,
	})
	if err != nil {
		return officeManager, errors.Wrap(err, "failed to update officeManager")
	}

	return officeManager, nil
}

func NewDefaultService() Service {
	return &OfficeManagerService{}
}

// ServiceMiddleware is a chainable behavior modifier for officeManagerResponsibleService.
type ServiceMiddleware func(OfficeManagerService) OfficeManagerService
