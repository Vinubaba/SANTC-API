package children

import (
	"context"
	"path"

	"github.com/Vinubaba/SANTC-API/common/firebase/claims"
	"github.com/Vinubaba/SANTC-API/common/storage"
	"github.com/Vinubaba/SANTC-API/common/store"

	. "github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/araddon/dateparse"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrNoParent         = errors.New("responsibleId is mandatory")
	ErrEmptyChild       = errors.New("childId cannot be empty")
	ErrDifferentDaycare = errors.New("child does not belong to this daycare")
	ErrUpdateDaycare    = errors.New("you can't update a child daycare")
)

type Service interface {
	AddChild(ctx context.Context, request ChildTransport) (store.Child, error)
	DeleteChild(ctx context.Context, request ChildTransport) error
	UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error)
	GetChild(ctx context.Context, request ChildTransport) (store.Child, error)
	ListChildren(ctx context.Context) ([]store.Child, error)

	AddPhoto(ctx context.Context, request PhotoRequestTransport) error
}

type ChildService struct {
	Store interface {
		Tx() *gorm.DB
		AddChild(tx *gorm.DB, child store.Child) (store.Child, error)
		UpdateChild(tx *gorm.DB, child store.Child) error
		GetChild(tx *gorm.DB, childId string, options store.SearchOptions) (store.Child, error)
		ListChildren(tx *gorm.DB, options store.SearchOptions) ([]store.Child, error)
		DeleteChild(tx *gorm.DB, childId string) error

		AddChildPhoto(tx *gorm.DB, childPhoto store.ChildPhoto) error

		GetClass(tx *gorm.DB, classId string, options store.SearchOptions) (store.Class, error)
		GetUser(tx *gorm.DB, userId string, searchOptions store.SearchOptions) (store.User, error)
	} `inject:""`
	Storage storage.Storage `inject:""`
	Logger  *log.Logger     `inject:""`
}

func (c *ChildService) AddChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	if IsNilOrEmpty(request.ResponsibleId) {
		return store.Child{}, ErrNoParent
	}

	if claims.IsAdmin(ctx) && IsNilOrEmpty(request.DaycareId) {
		return store.Child{}, errors.New("as an admin, you must specify the user daycare")
	} else {
		// default to requester daycare (e.g office manager)
		daycareId := claims.GetDaycareId(ctx)
		if IsNilOrEmpty(request.DaycareId) {
			request.DaycareId = &daycareId
		}

		if daycareId != *request.DaycareId {
			return store.Child{}, ErrDifferentDaycare
		}
	}

	if !IsNilOrEmpty(request.ImageUri) {
		imageUri, err := c.Storage.Store(ctx, *request.ImageUri, c.storageFolder(*request.DaycareId))
		if err != nil {
			return store.Child{}, errors.Wrap(err, "failed to store image")
		}
		request.ImageUri = &imageUri
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

	uri, err := c.Storage.Get(ctx, *request.ImageUri)
	if err != nil {
		tx.Rollback()
		return store.Child{}, errors.Wrap(err, "failed to generate image uri")
	}
	child.ImageUri = store.DbNullString(&uri)

	tx.Commit()
	return child, nil
}

func (c *ChildService) GetChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	if IsNilOrEmpty(request.Id) {
		return store.Child{}, ErrEmptyChild
	}

	child, err := c.Store.GetChild(nil, *request.Id, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return child, errors.Wrap(err, "failed to get child")
	}

	uri, err := c.Storage.Get(ctx, child.ImageUri.String)
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to generate image uri")
	}
	child.ImageUri = store.DbNullString(&uri)

	return child, nil
}

func (c *ChildService) DeleteChild(ctx context.Context, request ChildTransport) error {
	if IsNilOrEmpty(request.Id) {
		return ErrEmptyChild
	}
	child, err := c.Store.GetChild(nil, *request.Id, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return errors.Wrap(err, "failed to delete child")
	}

	if err := c.Store.DeleteChild(nil, *request.Id); err != nil {
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
		children[i].ImageUri = store.DbNullString(&uri)
	}

	return children, nil
}

func (c *ChildService) storageFolder(daycareId string) string {
	return path.Join("daycares", daycareId, "children")
}

func (c *ChildService) UpdateChild(ctx context.Context, request ChildTransport) (store.Child, error) {
	if IsNilOrEmpty(request.Id) {
		return store.Child{}, ErrEmptyChild
	}

	child, err := c.Store.GetChild(nil, *request.Id, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return store.Child{}, errors.Wrap(err, "failed to update child")
	}

	// User cannot update child daycare for the moment
	if !IsNilOrEmpty(request.DaycareId) && child.DaycareId.String != *request.DaycareId {
		return store.Child{}, ErrUpdateDaycare
	}

	if !IsNilOrEmpty(request.ImageUri) {
		imageUri, err := c.Storage.Store(ctx, *request.ImageUri, c.storageFolder(child.DaycareId.String))
		if err != nil {
			return store.Child{}, errors.Wrap(err, "failed to store image")
		}
		request.ImageUri = &imageUri
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
	uri, err := c.Storage.Get(ctx, child.ImageUri.String)
	if err != nil {
		c.Logger.Warn(ctx, "failed to generate image uri", "imageUri", child.ImageUri, "err", err.Error())
	}
	child.ImageUri = store.DbNullString(&uri)
}

func (c *ChildService) AddPhoto(ctx context.Context, request PhotoRequestTransport) error {
	child, err := c.GetChild(ctx, ChildTransport{Id: request.ChildId})
	if err != nil {
		return err
	}

	requesterUser, err := c.Store.GetUser(nil, *request.SenderId, store.SearchOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if child.DaycareId.String != requesterUser.DaycareId.String {
		return ErrDifferentDaycare
	}

	if err := c.Store.AddChildPhoto(nil, photoTransportToStore(request)); err != nil {
		return errors.Wrap(err, "failed to store photo")
	}

	return nil
}

func transportToStore(request ChildTransport, strict bool) (store.Child, error) {
	var birthDate, startDate time.Time
	var err error

	// In case of AddChild, dates are mandatory while in case of update they are not
	if strict || (!strict && !IsNilOrEmpty(request.BirthDate)) {
		birthDate, err = dateparse.ParseIn(*request.BirthDate, time.UTC)
		if err != nil {
			return store.Child{}, err
		}
	}

	if strict || (!strict && !IsNilOrEmpty(request.StartDate)) {
		startDate, err = dateparse.ParseIn(*request.StartDate, time.UTC)
		if err != nil {
			return store.Child{}, err
		}
	}

	child := store.Child{
		ChildId:       store.DbNullString(request.Id),
		DaycareId:     store.DbNullString(request.DaycareId),
		ClassId:       store.DbNullString(request.ClassId),
		ScheduleId:    store.DbNullString(request.Schedule.Id),
		BirthDate:     birthDate,
		FirstName:     store.DbNullString(request.FirstName),
		LastName:      store.DbNullString(request.LastName),
		Gender:        store.DbNullString(request.Gender),
		ImageUri:      store.DbNullString(request.ImageUri),
		Notes:         store.DbNullString(request.Notes),
		StartDate:     startDate,
		ResponsibleId: store.DbNullString(request.ResponsibleId),
		Relationship:  store.DbNullString(request.Relationship),
		Schedule: store.Schedule{
			ScheduleId:     store.DbNullString(request.Schedule.Id),
			WalkIn:         store.DbNullBool(request.Schedule.WalkIn),
			MondayStart:    store.DbNullString(request.Schedule.MondayStart),
			MondayEnd:      store.DbNullString(request.Schedule.MondayEnd),
			TuesdayStart:   store.DbNullString(request.Schedule.TuesdayStart),
			TuesdayEnd:     store.DbNullString(request.Schedule.TuesdayEnd),
			WednesdayStart: store.DbNullString(request.Schedule.WednesdayStart),
			WednesdayEnd:   store.DbNullString(request.Schedule.WednesdayEnd),
			ThursdayStart:  store.DbNullString(request.Schedule.ThursdayStart),
			ThursdayEnd:    store.DbNullString(request.Schedule.ThursdayEnd),
			FridayStart:    store.DbNullString(request.Schedule.FridayStart),
			FridayEnd:      store.DbNullString(request.Schedule.FridayEnd),
			SaturdayStart:  store.DbNullString(request.Schedule.SaturdayStart),
			SaturdayEnd:    store.DbNullString(request.Schedule.SaturdayEnd),
			SundayStart:    store.DbNullString(request.Schedule.SundayStart),
			SundayEnd:      store.DbNullString(request.Schedule.SundayEnd),
		},
	}
	for _, specialInstruction := range request.SpecialInstructions {
		instructionToCreate := store.SpecialInstruction{Instruction: store.DbNullString(specialInstruction.Instruction)}
		child.SpecialInstructions = append(child.SpecialInstructions, instructionToCreate)
	}
	for _, allergy := range request.Allergies {
		allergyToCreate := store.Allergy{Allergy: store.DbNullString(allergy.Allergy), Instruction: store.DbNullString(allergy.Instruction)}
		child.Allergies = append(child.Allergies, allergyToCreate)
	}
	return child, nil
}

func photoTransportToStore(request PhotoRequestTransport) store.ChildPhoto {
	childPhoto := store.ChildPhoto{
		ChildId:         store.DbNullString(request.ChildId),
		ImageUri:        store.DbNullString(request.Filename),
		Approved:        false,
		PublishedBy:     store.DbNullString(request.SenderId),
		PublicationDate: time.Now(),
	}
	return childPhoto
}

// ServiceMiddleware is a chainable behavior modifier for childService.
type ServiceMiddleware func(ChildService) ChildService
