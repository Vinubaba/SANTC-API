package adult_responsible

import (
	"arthurgustin.fr/teddycare/store"
	"context"
	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
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
		AddUser(tx *gorm.DB, user store.User) (id string, err error)
		AddAdultResponsible(tx *gorm.DB, adult store.AdultResponsible) (store.AdultResponsible, error)
		ListAdultResponsible(tx *gorm.DB) ([]store.AdultResponsible, error)
		DeleteAdultResponsible(tx *gorm.DB, adultId string) error
		GetAdultResponsible(tx *gorm.DB, adultId string) (store.AdultResponsible, error)
		UpdateAdultResponsible(tx *gorm.DB, adult store.AdultResponsible) (store.AdultResponsible, error)
		AddRole(tx *gorm.DB, role store.Role) (store.Role, error)
		Tx() *gorm.DB
	} `inject:""`
}

func (c *AdultResponsibleService) AddAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error) {
	if err := c.validateAddAdultResponsibleRequest(request); err != nil {
		return store.AdultResponsible{}, err
	}

	tx := c.Store.Tx()

	userId, err := c.Store.AddUser(tx, store.User{
		Password: request.Password,
		Email:    request.Email,
	})
	if err != nil {
		tx.Rollback()
		return store.AdultResponsible{}, errors.New("failed to create user")
	}

	adult, err := c.Store.AddAdultResponsible(tx, store.AdultResponsible{
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
		tx.Rollback()
		return store.AdultResponsible{}, errors.Wrap(err, "failed to add adult")
	}

	_, err = c.Store.AddRole(tx, store.Role{
		UserId: adult.ResponsibleId,
		Role:   store.ROLE_ADULT,
	})
	if err != nil {
		tx.Rollback()
		return store.AdultResponsible{}, errors.Wrap(err, "failed to add role")
	}

	tx.Commit()
	return adult, nil
}

func (c *AdultResponsibleService) validateAddAdultResponsibleRequest(req AdultResponsibleTransport) error {
	if err := checkmail.ValidateFormat(req.Email); err != nil {
		return ErrInvalidEmail
	}
	if len(req.Password) < 6 {
		return ErrInvalidPasswordFormat
	}
	return nil
}

func (c *AdultResponsibleService) ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error) {
	adults, err := c.Store.ListAdultResponsible(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list adult")
	}

	return adults, nil
}

func (c AdultResponsibleService) DeleteAdultResponsible(ctx context.Context, request AdultResponsibleTransport) error {
	if err := c.Store.DeleteAdultResponsible(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete adult")
	}

	return nil
}

func (c *AdultResponsibleService) GetAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error) {
	adult, err := c.Store.GetAdultResponsible(nil, request.Id)
	if err != nil {
		return adult, errors.Wrap(err, "failed to get adult")
	}

	return adult, nil
}

func (c *AdultResponsibleService) UpdateAdultResponsible(ctx context.Context, request AdultResponsibleTransport) (store.AdultResponsible, error) {
	if request.Email != "" && checkmail.ValidateFormat(request.Email) != nil {
		return store.AdultResponsible{}, ErrInvalidEmail
	}

	adult, err := c.Store.UpdateAdultResponsible(nil, store.AdultResponsible{
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

// ServiceMiddleware is a chainable behavior modifier for adultResponsibleService.
type ServiceMiddleware func(AdultResponsibleService) AdultResponsibleService
