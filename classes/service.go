package classes

import (
	"context"
	"strings"

	"github.com/Vinubaba/SANTC-API/storage"
	"github.com/Vinubaba/SANTC-API/store"

	"github.com/Vinubaba/SANTC-API/ageranges"
	"github.com/Vinubaba/SANTC-API/shared"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrEmptyClass    = errors.New("classId cannot be empty")
	ErrInvalidImage  = errors.New("for now, only jpeg is supported. the image must have the following pattern: 'data:image/jpeg;base64,[big 64encoded image string]'")
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

	if (ageranges.AgeRangeTransport{}) == request.AgeRange {
		return store.Class{}, ErrEmptyAgeRange
	}

	var filename string
	if request.ImageUri != "" {
		encoded, mimeType, err := c.validate64EncodedPhoto(request.ImageUri)
		if err != nil {
			return store.Class{}, errors.Wrap(err, "failed to validate image")
		}

		filename, err = c.Storage.Store(ctx, encoded, mimeType)
		if err != nil {
			return store.Class{}, errors.Wrap(err, "failed to store image")
		}
	}

	tx := c.Store.Tx()
	if tx.Error != nil {
		return store.Class{}, errors.Wrap(tx.Error, "failed to add class")
	}

	class, err := c.Store.AddClass(tx, store.Class{
		Name:        store.DbNullString(request.Name),
		Description: store.DbNullString(request.Description),
		ImageUri:    store.DbNullString(filename),
		AgeRangeId:  store.DbNullString(request.AgeRange.Id),
		AgeRange: store.AgeRange{
			AgeRangeId: store.DbNullString(request.AgeRange.Id),
			Stage:      store.DbNullString(request.AgeRange.Stage),
			Min:        request.AgeRange.Min,
			MinUnit:    store.DbNullString(request.AgeRange.MinUnit),
			Max:        request.AgeRange.Max,
			MaxUnit:    store.DbNullString(request.AgeRange.MaxUnit),
		},
	})
	if err != nil {
		tx.Rollback()
		return store.Class{}, errors.Wrap(err, "failed to add class")
	}

	uri, err := c.Storage.Get(ctx, filename)
	if err != nil {
		tx.Rollback()
		return store.Class{}, errors.Wrap(err, "failed to generate image uri")
	}
	// When adding a class, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
	class.ImageUri = store.DbNullString(uri)

	tx.Commit()
	return class, nil
}

func (c *ClassService) validate64EncodedPhoto(photo string) (encoded, mimeType string, err error) {
	if strings.HasPrefix(photo, "data:image/jpeg;base64,") {
		mimeType = "image/jpeg"
		encoded = strings.TrimPrefix(photo, "data:image/jpeg;base64,")
	} else {
		err = ErrInvalidImage
	}
	return
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
	// When adding a class, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
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
		// When adding a class, the json response will contains a temporary uri, so the frontend can do whatever it wants with it
		classes[i].ImageUri = store.DbNullString(uri)
	}

	return classes, nil
}

func (c *ClassService) UpdateClass(ctx context.Context, request ClassTransport) (store.Class, error) {
	if request.Id == "" {
		return store.Class{}, ErrEmptyClass
	}

	if _, err := c.Store.GetClass(nil, request.Id); err != nil {
		return store.Class{}, errors.Wrap(err, "failed to update class")
	}

	if err := c.setAndStoreDecodedImage(ctx, &request); err != nil {
		return store.Class{}, err
	}

	tx := c.Store.Tx()
	if tx.Error != nil {
		return store.Class{}, errors.Wrap(tx.Error, "failed to update class")
	}

	class, err := c.Store.UpdateClass(tx, store.Class{
		ImageUri:    store.DbNullString(request.ImageUri),
		ClassId:     store.DbNullString(request.Id),
		Description: store.DbNullString(request.Description),
		Name:        store.DbNullString(request.Name),
		AgeRangeId:  store.DbNullString(request.AgeRange.Id),
		AgeRange: store.AgeRange{
			AgeRangeId: store.DbNullString(request.AgeRange.Id),
			Stage:      store.DbNullString(request.AgeRange.Stage),
			Min:        request.AgeRange.Min,
			MinUnit:    store.DbNullString(request.AgeRange.MinUnit),
			Max:        request.AgeRange.Max,
			MaxUnit:    store.DbNullString(request.AgeRange.MaxUnit),
		},
	})
	if err != nil {
		tx.Rollback()
		return class, errors.Wrap(err, "failed to update class")
	}

	tx.Commit()
	c.setBucketUri(ctx, &class)
	return class, nil
}

func (c *ClassService) setAndStoreDecodedImage(ctx context.Context, request *ClassTransport) error {
	if strings.HasPrefix(request.ImageUri, "data:image/jpeg;base64,") {
		mimeType := "image/jpeg"
		encoded := strings.TrimPrefix(request.ImageUri, "data:image/jpeg;base64,")

		var err error
		request.ImageUri, err = c.Storage.Store(ctx, encoded, mimeType)
		if err != nil {
			return errors.Wrap(err, "failed to store image")
		}
	}
	return nil
}

func (c *ClassService) setBucketUri(ctx context.Context, class *store.Class) {
	if class.ImageUri.String == "" {
		return
	}
	if !strings.Contains(class.ImageUri.String, "/") {
		uri, err := c.Storage.Get(ctx, class.ImageUri.String)
		if err != nil {
			c.Logger.Warn(ctx, "failed to generate image uri", "imageUri", class.ImageUri, "err", err.Error())
		}
		class.ImageUri = store.DbNullString(uri)
	}
}

// ServiceMiddleware is a chainable behavior modifier for classService.
type ServiceMiddleware func(ClassService) ClassService
