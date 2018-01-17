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
	AddChild(ctx context.Context, request ChildRequest) (string, error)
}

type ChildService struct {
	Store interface {
		BeginTransaction()
		Commit()
		Rollback()
		AddChild(ctx context.Context, child store.Child) (id string, err error)
		SetResponsible(ctx context.Context, responsibleOf store.ResponsibleOf) error
	} `inject:""`
}

func (c ChildService) AddChild(ctx context.Context, request ChildRequest) (string, error) {
	t, err := dateparse.ParseIn(request.BirthDate, time.UTC)
	if err != nil {
		return "", err
	}

	if request.ResponsibleId == "" {
		return "", ErrNoParent
	}

	c.Store.BeginTransaction()

	childId, err := c.Store.AddChild(ctx, store.Child{
		BirthDate: t,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	})
	if err != nil {
		c.Store.Rollback()
		return "", errors.Wrap(err, "failed to add child")
	}

	err = c.Store.SetResponsible(ctx, store.ResponsibleOf{Relationship: request.Relationship, ChildId: childId, ResponsibleId: request.ResponsibleId})
	if err != nil {
		c.Store.Rollback()
		return "", errors.Wrap(err, "failed to set responsible")
	}
	c.Store.Commit()
	return childId, nil
}

// ServiceMiddleware is a chainable behavior modifier for childService.
type ServiceMiddleware func(ChildService) ChildService
