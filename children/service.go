package children

import (
	"arthurgustin.fr/teddycare/store"
	"context"
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrNoParent       = errors.New("responsibleId is mandatory")
	ErrEmptyChild     = errors.New("childId cannot be empty")
	ErrSetResponsible = errors.New("failed to set responsibleId")
	ErrSetAllergy     = errors.New("failed to set allergy")
)

type Service interface {
	AddChild(ctx context.Context, request ChildTransport) (store.Child, error)
	DeleteChild(ctx context.Context, request ChildTransport) error
	UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error)
	GetChild(ctx context.Context, request ChildTransport) (store.Child, error)
	ListChild(ctx context.Context) ([]store.Child, error)
	FindAllergiesOfChild(ctx context.Context, childId string) ([]store.Allergy, error)
}

type ChildService struct {
	Store interface {
		BeginTransaction()
		Commit()
		Rollback()
		AddChild(context.Context, store.Child) (store.Child, error)
		UpdateChild(context.Context, store.Child) (store.Child, error)
		GetChild(ctx context.Context, childId string) (store.Child, error)
		ListChild(context.Context) ([]store.Child, error)
		DeleteChild(ctx context.Context, childId string) error
		SetResponsible(ctx context.Context, responsibleOf store.ResponsibleOf) error

		AddAllergy(ctx context.Context, allergy store.Allergy) (store.Allergy, error)
		FindAllergiesOfChild(ctx context.Context, childId string) ([]store.Allergy, error)
		RemoveAllergiesOfChild(ctx context.Context, childId string) error
	} `inject:""`
}

func (c ChildService) AddChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	t, err := dateparse.ParseIn(request.BirthDate, time.UTC)
	if err != nil {
		return store.Child{}, err
	}

	if request.ResponsibleId == "" {
		return store.Child{}, ErrNoParent
	}

	c.Store.BeginTransaction()

	child, err := c.Store.AddChild(ctx, store.Child{
		BirthDate:   t,
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Gender:      request.Gender,
		PicturePath: request.PicturePath,
	})
	if err != nil {
		c.Store.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to add child")
	}

	if err = c.Store.SetResponsible(ctx, store.ResponsibleOf{Relationship: request.Relationship, ChildId: child.ChildId, ResponsibleId: request.ResponsibleId}); err != nil {
		c.Store.Rollback()
		return store.Child{}, errors.Wrap(ErrSetResponsible, "failed to set responsible. err: "+err.Error())
	}

	for _, allergy := range request.Allergies {
		if _, err := c.Store.AddAllergy(ctx, store.Allergy{ChildId: child.ChildId, Allergy: allergy}); err != nil {
			c.Store.Rollback()
			return store.Child{}, errors.Wrap(ErrSetAllergy, "failed to set allergy. err: "+err.Error())
		}
	}

	c.Store.Commit()
	return child, nil
}

func (c ChildService) GetChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	adult, err := c.Store.GetChild(ctx, request.Id)
	if err != nil {
		return adult, errors.Wrap(err, "failed to get child")
	}

	return adult, nil
}

func (c ChildService) DeleteChild(ctx context.Context, request ChildTransport) error {
	if err := c.Store.DeleteChild(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete child")
	}

	return nil
}

func (c ChildService) ListChild(ctx context.Context) ([]store.Child, error) {
	children, err := c.Store.ListChild(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list adult")
	}

	return children, nil
}

func (c ChildService) UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	var t time.Time
	var err error

	if request.Id == "" {
		return store.Child{}, ErrEmptyChild
	}

	if request.BirthDate != "" {
		t, err = dateparse.ParseIn(request.BirthDate, time.UTC)
		if err != nil {
			return store.Child{}, err
		}
	}

	c.Store.BeginTransaction()

	if len(request.Allergies) > 1 {
		if err := c.Store.RemoveAllergiesOfChild(ctx, request.Id); err != nil {
			c.Store.Rollback()
			return store.Child{}, errors.Wrap(err, "failed to delete allergies")
		}
		for _, allergy := range request.Allergies {
			if _, err := c.Store.AddAllergy(ctx, store.Allergy{ChildId: request.Id, Allergy: allergy}); err != nil {
				c.Store.Rollback()
				return store.Child{}, errors.Wrap(ErrSetAllergy, "failed to set allergy. err: "+err.Error())
			}
		}
	}

	child, err := c.Store.UpdateChild(ctx, store.Child{
		BirthDate:   t,
		PicturePath: request.PicturePath,
		Gender:      request.Gender,
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		ChildId:     request.Id,
	})
	if err != nil {
		c.Store.Rollback()
		return child, errors.Wrap(err, "failed to update child")
	}

	c.Store.Commit()

	return child, nil
}

func (c ChildService) FindAllergiesOfChild(ctx context.Context, childId string) ([]store.Allergy, error) {
	allergies, err := c.Store.FindAllergiesOfChild(ctx, childId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find allergies")
	}

	return allergies, nil
}

// ServiceMiddleware is a chainable behavior modifier for childService.
type ServiceMiddleware func(ChildService) ChildService
