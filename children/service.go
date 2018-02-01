package children

import (
	"context"
	"strings"
	"time"

	"arthurgustin.fr/teddycare/storage"
	"arthurgustin.fr/teddycare/store"

	"github.com/araddon/dateparse"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
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
		Tx() *gorm.DB
		AddChild(tx *gorm.DB, child store.Child) (store.Child, error)
		UpdateChild(tx *gorm.DB, child store.Child) (store.Child, error)
		GetChild(tx *gorm.DB, childId string) (store.Child, error)
		ListChild(tx *gorm.DB) ([]store.Child, error)
		DeleteChild(tx *gorm.DB, childId string) error
		SetResponsible(tx *gorm.DB, responsibleOf store.ResponsibleOf) error

		AddAllergy(tx *gorm.DB, allergy store.Allergy) (store.Allergy, error)
		FindAllergiesOfChild(tx *gorm.DB, childId string) ([]store.Allergy, error)
		RemoveAllergiesOfChild(tx *gorm.DB, childId string) error
	} `inject:""`
	Storage storage.Storage `inject:""`
}

func (c *ChildService) AddChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	t, err := dateparse.ParseIn(request.BirthDate, time.UTC)
	if err != nil {
		return store.Child{}, err
	}

	if request.ResponsibleId == "" {
		return store.Child{}, ErrNoParent
	}

	encoded, mimeType, err := c.validate64EncodedPhoto(request.Image)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to validate image")
	}

	filename, err := c.Storage.Store(context.Background(), encoded, mimeType)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to store image")
	}

	tx := c.Store.Tx()

	child, err := c.Store.AddChild(tx, store.Child{
		BirthDate: t,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Gender:    request.Gender,
		ImageUri:  filename,
	})
	if err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to add child")
	}

	uri, err := c.Storage.Get(context.Background(), filename)
	if err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to generate image uri")
	}
	// When adding a child, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
	child.ImageUri = uri

	if err = c.Store.SetResponsible(tx, store.ResponsibleOf{Relationship: request.Relationship, ChildId: child.ChildId, ResponsibleId: request.ResponsibleId}); err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(ErrSetResponsible, "failed to set responsible. err: "+err.Error())
	}

	for _, allergy := range request.Allergies {
		if _, err := c.Store.AddAllergy(tx, store.Allergy{ChildId: child.ChildId, Allergy: allergy}); err != nil {
			tx.Rollback()
			return store.Child{}, errors.Wrap(ErrSetAllergy, "failed to set allergy. err: "+err.Error())
		}
	}

	tx.Commit()
	return child, nil
}

var (
	ErrInvalidImage = errors.New("for now, only jpeg is supported. the image must have the following pattern: 'data:image/jpeg;base64,[big 64encoded image string]'")
)

func (c *ChildService) validate64EncodedPhoto(photo string) (encoded, mimeType string, err error) {
	if strings.HasPrefix(photo, "data:image/jpeg;base64,") {
		mimeType = "image/jpeg"
		encoded = strings.TrimPrefix(photo, "data:image/jpeg;base64,")
	} else {
		err = ErrInvalidImage
	}
	return
}

func (c *ChildService) GetChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	child, err := c.Store.GetChild(nil, request.Id)
	if err != nil {
		return child, errors.Wrap(err, "failed to get child")
	}

	uri, err := c.Storage.Get(ctx, child.ImageUri)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to generate image uri")
	}
	// When adding a child, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
	child.ImageUri = uri

	return child, nil
}

func (c *ChildService) DeleteChild(ctx context.Context, request ChildTransport) error {
	child, err := c.Store.GetChild(nil, request.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get child")
	}

	if err := c.Store.DeleteChild(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete child")
	}

	if err := c.Storage.Delete(ctx, child.ImageUri); err != nil {
		return errors.Wrap(err, "failed to delete child image")
	}

	return nil
}

func (c *ChildService) ListChild(ctx context.Context) ([]store.Child, error) {
	children, err := c.Store.ListChild(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list adult")
	}

	for i := 0; i < len(children); i++ {
		uri, err := c.Storage.Get(ctx, children[i].ImageUri)
		if err != nil {
			return []store.Child{}, errors.Wrap(err, "failed to generate image uri")
		}
		// When adding a child, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
		children[i].ImageUri = uri
	}

	return children, nil
}

func (c *ChildService) UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error) {
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

	tx := c.Store.Tx()

	if len(request.Allergies) > 1 {
		if err := c.Store.RemoveAllergiesOfChild(tx, request.Id); err != nil {
			tx.Rollback()
			return store.Child{}, errors.Wrap(err, "failed to delete allergies")
		}
		for _, allergy := range request.Allergies {
			if _, err := c.Store.AddAllergy(tx, store.Allergy{ChildId: request.Id, Allergy: allergy}); err != nil {
				tx.Rollback()
				return store.Child{}, errors.Wrap(ErrSetAllergy, "failed to set allergy. err: "+err.Error())
			}
		}
	}

	child, err := c.Store.UpdateChild(tx, store.Child{
		BirthDate: t,
		ImageUri:  request.Image,
		Gender:    request.Gender,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		ChildId:   request.Id,
	})
	if err != nil {
		tx.Rollback()
		return child, errors.Wrap(err, "failed to update child")
	}

	tx.Commit()

	return child, nil
}

func (c *ChildService) FindAllergiesOfChild(ctx context.Context, childId string) ([]store.Allergy, error) {
	allergies, err := c.Store.FindAllergiesOfChild(nil, childId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find allergies")
	}

	return allergies, nil
}

// ServiceMiddleware is a chainable behavior modifier for childService.
type ServiceMiddleware func(ChildService) ChildService
