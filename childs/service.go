package childs

import (
	"arthurgustin.fr/teddycare/store"
	"context"
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrNoParent = errors.New("responsibleId is mandatory")
)

type Service interface {
	AddChild(ctx context.Context, request ChildRequest) (store.Child, error)
}

type ChildService struct {
	Store interface {
		BeginTransaction()
		Commit()
		Rollback()
		AddChild(context.Context, store.Child) (store.Child, error)
		SetResponsible(ctx context.Context, responsibleOf store.ResponsibleOf) error
	} `inject:""`
}

func (c ChildService) AddChild(ctx context.Context, request ChildRequest) (store.Child, error) {
	t, err := dateparse.ParseIn(request.BirthDate, time.UTC)
	if err != nil {
		return store.Child{}, err
	}

	if request.ResponsibleId == "" {
		return store.Child{}, ErrNoParent
	}

	c.Store.BeginTransaction()

	child, err := c.Store.AddChild(ctx, store.Child{
		BirthDate: t,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	})
	if err != nil {
		c.Store.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to add child")
	}

	err = c.Store.SetResponsible(ctx, store.ResponsibleOf{Relationship: request.Relationship, ChildId: child.ChildId, ResponsibleId: request.ResponsibleId})
	if err != nil {
		c.Store.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to set responsible")
	}
	c.Store.Commit()
	return child, nil
}

// ServiceMiddleware is a chainable behavior modifier for childService.
type ServiceMiddleware func(ChildService) ChildService
