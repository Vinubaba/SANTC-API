package classes

import (
	"context"
	"strings"

	"github.com/Vinubaba/SANTC-API/storage"
	"github.com/Vinubaba/SANTC-API/store"

	"database/sql"
	"github.com/Vinubaba/SANTC-API/ageranges"
	"github.com/Vinubaba/SANTC-API/shared"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrEmptyClass    = errors.New("classId cannot be empty")
	ErrEmptyAgeRange = errors.New("please specify an age range")
)

type Service interface {
	AddClass(ctx context.Context, request ClassTransport) (store.Class, error)
	DeleteClass(ctx context.Context, request ClassTransport) error
	UpdateClass(ctx context.Context, request ClassTransport) (store.Class, error)
	GetClass(ctx context.Context, request ClassTransport) (store.Class, error)
	ListClass(ctx context.Context) ([]store.Class, error)
}

type ClassService struct {
	Store interface {
		Tx() *gorm.DB

		AddClass(tx *gorm.DB, class store.Class) (store.Class, error)
		UpdateClass(tx *gorm.DB, class store.Class) (store.Class, error)
		GetClass(tx *gorm.DB, classId string) (store.Class, error)
		ListClass(tx *gorm.DB) ([]store.Class, error)
		DeleteClass(tx *gorm.DB, classId string) error
	} `inject:""`
	Storage storage.Storage `inject:""`
	Logger  *shared.Logger  `inject:""`
}

func (c *ClassService) AddClass(ctx context.Context, request ClassTransport) (store.Class, error) {
	var err error

	if (ageranges.AgeRangeTransport{}) == request.AgeRange {
		return store.Class{}, ErrEmptyAgeRange
	}

	request.ImageUri, err = c.Storage.Store(ctx, request.ImageUri)
	if err != nil {
		return store.Class{}, errors.Wrap(err, "failed to store image")
	}

	class, err := c.Store.AddClass(nil, transportToStore(request))
	if err != nil {
		return store.Class{}, errors.Wrap(err, "failed to add class")
	}

	uri, err := c.Storage.Get(ctx, request.ImageUri)
	if err != nil {
		return store.Class{}, errors.Wrap(err, "failed to generate image uri")
	}
	class.ImageUri = store.DbNullString(uri)

	return class, nil
}

func (c *ClassService) GetClass(ctx context.Context, request ClassTransport) (store.Class, error) {
	class, err := c.Store.GetClass(nil, request.Id)
	if err != nil {
		return class, errors.Wrap(err, "failed to get class")
	}

	uri, err := c.Storage.Get(ctx, class.ImageUri.String)
	if err != nil {
		return store.Class{}, errors.Wrap(err, "failed to generate image uri")
	}
	class.ImageUri = store.DbNullString(uri)

	return class, nil
}

func (c *ClassService) DeleteClass(ctx context.Context, request ClassTransport) error {
	class, err := c.Store.GetClass(nil, request.Id)
	if err != nil {
		return errors.Wrap(err, "failed to delete class")
	}

	if err := c.Store.DeleteClass(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete class")
	}

	if err := c.Storage.Delete(ctx, class.ImageUri.String); err != nil {
		return errors.Wrap(err, "failed to delete class image")
	}

	return nil
}

func (c *ClassService) ListClass(ctx context.Context) ([]store.Class, error) {
	classes, err := c.Store.ListClass(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list classes")
	}

	for i := 0; i < len(classes); i++ {
		uri, err := c.Storage.Get(ctx, classes[i].ImageUri.String)
		if err != nil {
			return []store.Class{}, errors.Wrap(err, "failed to generate image uri")
		}
		classes[i].ImageUri = store.DbNullString(uri)
	}

	return classes, nil
}

func (c *ClassService) UpdateClass(ctx context.Context, request ClassTransport) (store.Class, error) {
	var err error

	if request.Id == "" {
		return store.Class{}, ErrEmptyClass
	}

	if _, err := c.Store.GetClass(nil, request.Id); err != nil {
		return store.Class{}, errors.Wrap(err, "failed to update class")
	}

	request.ImageUri, err = c.Storage.Store(ctx, request.ImageUri)
	if err != nil {
		return store.Class{}, errors.Wrap(err, "failed to store image")
	}

	class, err := c.Store.UpdateClass(nil, transportToStore(request))
	if err != nil {
		return class, errors.Wrap(err, "failed to update class")
	}

	uri, err := c.Storage.Get(ctx, request.ImageUri)
	if err != nil {
		return store.Class{}, errors.Wrap(err, "failed to generate image uri")
	}
	class.ImageUri = store.DbNullString(uri)
	return class, nil
}

func (c *ClassService) getBucketUri(ctx context.Context, imgPath string) sql.NullString {
	if imgPath == "" || strings.Contains(imgPath, "/") {
		return sql.NullString{
			String: "",
			Valid:  false,
		}
	}
	uri, err := c.Storage.Get(ctx, imgPath)
	if err != nil {
		c.Logger.Warn(ctx, "failed to generate image uri", "imageUri", imgPath, "err", err.Error())
	}
	return store.DbNullString(uri)
}

// ServiceMiddleware is a chainable behavior modifier for classService.
type ServiceMiddleware func(ClassService) ClassService
