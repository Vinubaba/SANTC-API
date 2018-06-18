package children

import (
	"context"
	"strings"

	"github.com/Vinubaba/SANTC-API/common/firebase/claims"
	"github.com/Vinubaba/SANTC-API/common/storage"
	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrNoParent               = errors.New("responsibleId is mandatory")
	ErrEmptyChild             = errors.New("childId cannot be empty")
	ErrInvalidImage           = errors.New("for now, only jpeg is supported. the image must have the following pattern: 'data:image/jpeg;base64,[big 64encoded image string]'")
	ErrCreateDifferentDaycare = errors.New("you can't add a child to a different daycare of you")
	ErrUpdateDaycare          = errors.New("you can't update a child daycare")
)

type Service interface {
	AddChild(ctx context.Context, request ChildTransport) (store.Child, error)
	DeleteChild(ctx context.Context, request ChildTransport) error
	UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error)
	GetChild(ctx context.Context, request ChildTransport) (store.Child, error)
	ListChildren(ctx context.Context) ([]store.Child, error)
}

type ChildService struct {
	Store interface {
		Tx() *gorm.DB
		AddChild(tx *gorm.DB, child store.Child) (store.Child, error)
		UpdateChild(tx *gorm.DB, child store.Child) error
		GetChild(tx *gorm.DB, childId string, options store.SearchOptions) (store.Child, error)
		ListChildren(tx *gorm.DB, options store.SearchOptions) ([]store.Child, error)
		DeleteChild(tx *gorm.DB, childId string) error

		GetClass(tx *gorm.DB, classId string, options store.SearchOptions) (store.Class, error)
	} `inject:""`
	Storage storage.Storage `inject:""`
	Logger  *log.Logger     `inject:""`
}

func (c *ChildService) AddChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	if request.ResponsibleId == "" {
		return store.Child{}, ErrNoParent
	}

	if claims.IsAdmin(ctx) && request.DaycareId == "" {
		return store.Child{}, errors.New("as an admin, you must specify the user daycare")
	} else {
		// default to requester daycare (e.g office manager)
		if request.DaycareId == "" {
			request.DaycareId = claims.GetDaycareId(ctx)
		}

		if claims.GetDaycareId(ctx) != request.DaycareId {
			return store.Child{}, ErrCreateDifferentDaycare
		}
	}

	var err error
	request.ImageUri, err = c.Storage.Store(ctx, request.ImageUri)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to store image")
	}

	tx := c.Store.Tx()
	if tx.Error != nil {
		return store.Child{}, errors.Wrap(tx.Error, "failed to add child")
	}

	childToCreate, err := transportToStore(request, true)
	if err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to decode request")
	}

	if childToCreate.ClassId.String != "" {
		_, err := c.Store.GetClass(nil, childToCreate.ClassId.String, store.SearchOptions{DaycareId: childToCreate.DaycareId.String})
		if err != nil {
			tx.Rollback()
			return store.Child{}, errors.Wrap(err, "failed to add child")
		}
	}

	child, err := c.Store.AddChild(tx, childToCreate)
	if err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to add child")
	}

	uri, err := c.Storage.Get(ctx, request.ImageUri)
	if err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to generate image uri")
	}
	child.ImageUri = store.DbNullString(uri)

	tx.Commit()
	return child, nil
}

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
	child, err := c.Store.GetChild(nil, request.Id, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return child, errors.Wrap(err, "failed to get child")
	}

	uri, err := c.Storage.Get(ctx, child.ImageUri.String)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to generate image uri")
	}
	child.ImageUri = store.DbNullString(uri)

	return child, nil
}

func (c *ChildService) DeleteChild(ctx context.Context, request ChildTransport) error {
	child, err := c.Store.GetChild(nil, request.Id, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return errors.Wrap(err, "failed to delete child")
	}

	if err := c.Store.DeleteChild(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete child")
	}

	if err := c.Storage.Delete(ctx, child.ImageUri.String); err != nil {
		return errors.Wrap(err, "failed to delete child image")
	}

	return nil
}

func (c *ChildService) ListChildren(ctx context.Context) ([]store.Child, error) {
	children, err := c.Store.ListChildren(nil, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to list children")
	}

	for i := 0; i < len(children); i++ {
		uri, err := c.Storage.Get(ctx, children[i].ImageUri.String)
		if err != nil {
			return []store.Child{}, errors.Wrap(err, "failed to generate image uri")
		}
		// When adding a child, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
		children[i].ImageUri = store.DbNullString(uri)
	}

	return children, nil
}

func (c *ChildService) UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	var err error

	if request.Id == "" {
		return store.Child{}, ErrEmptyChild
	}

	child, err := c.Store.GetChild(nil, request.Id, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to update child")
	}

	// User cannot update child daycare for the moment
	if request.DaycareId != "" && child.DaycareId.String != request.DaycareId {
		return store.Child{}, ErrUpdateDaycare
	}

	request.ImageUri, err = c.Storage.Store(ctx, request.ImageUri)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to store image")
	}

	childToUpdate, err := transportToStore(request, false)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to decode request")
	}

	err = c.Store.UpdateChild(nil, childToUpdate)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to update child")
	}

	childToReturn, err := c.Store.GetChild(nil, childToUpdate.ChildId.String, store.SearchOptions{})
	if err != nil {
		return store.Child{}, err
	}
	c.setBucketUri(ctx, &childToReturn)
	return childToReturn, nil
}

func (c *ChildService) setBucketUri(ctx context.Context, child *store.Child) {
	if child.ImageUri.String == "" {
		return
	}
	if !strings.Contains(child.ImageUri.String, "/") {
		uri, err := c.Storage.Get(ctx, child.ImageUri.String)
		if err != nil {
			c.Logger.Warn(ctx, "failed to generate image uri", "imageUri", child.ImageUri, "err", err.Error())
		}
		child.ImageUri = store.DbNullString(uri)
	}
}

// ServiceMiddleware is a chainable behavior modifier for childService.
type ServiceMiddleware func(ChildService) ChildService
