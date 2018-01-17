package adult_responsible

import (
	"arthurgustin.fr/teddycare/store"
	"context"
	"github.com/pkg/errors"
)

type Service interface {
	AddAdultResponsible(ctx context.Context, request AdultResponsibleRequest) (string, error)
	ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error)
}

type AdultResponsibleService struct {
	Store interface {
		AddAdultResponsible(ctx context.Context, adult store.AdultResponsible) (id string, err error)
		ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error)
	} `inject:""`
}

func (c AdultResponsibleService) AddAdultResponsible(ctx context.Context, request AdultResponsibleRequest) (string, error) {
	responsibleId, err := c.Store.AddAdultResponsible(ctx, store.AdultResponsible{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Gender:    request.Gender,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to add adult")
	}

	return responsibleId, nil
}

func (c AdultResponsibleService) ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error) {
	adults, err := c.Store.ListAdultResponsible(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list adult")
	}

	return adults, nil
}

func NewDefaultService() Service {
	return &AdultResponsibleService{}
}

// ServiceMiddleware is a chainable behavior modifier for adultResponsibleService.
type ServiceMiddleware func(AdultResponsibleService) AdultResponsibleService
