package ageranges

import (
	"context"

	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/Vinubaba/SANTC-API/common/firebase/claims"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrEmptyAgeRange = errors.New("ageRangeId cannot be empty")
)

type Service interface {
	AddAgeRange(ctx context.Context, request AgeRangeTransport) (store.AgeRange, error)
	DeleteAgeRange(ctx context.Context, request AgeRangeTransport) error
	UpdateAgeRange(ctx context.Context, request AgeRangeTransport) (store.AgeRange, error)
	GetAgeRange(ctx context.Context, request AgeRangeTransport) (store.AgeRange, error)
	ListAgeRange(ctx context.Context) ([]store.AgeRange, error)
}

type AgeRangeService struct {
	Store interface {
		Tx() *gorm.DB

		AddAgeRange(tx *gorm.DB, ageRange store.AgeRange) (store.AgeRange, error)
		UpdateAgeRange(tx *gorm.DB, ageRange store.AgeRange) (store.AgeRange, error)
		GetAgeRange(tx *gorm.DB, ageRangeId string, options store.SearchOptions) (store.AgeRange, error)
		ListAgeRange(tx *gorm.DB, options store.SearchOptions) ([]store.AgeRange, error)
		DeleteAgeRange(tx *gorm.DB, ageRangeId string) error
	} `inject:""`
	Logger *log.Logger `inject:""`
}

func (c *AgeRangeService) transportToStore(request AgeRangeTransport) store.AgeRange {
	return store.AgeRange{
		AgeRangeId: store.DbNullString(request.Id),
		DaycareId:  store.DbNullString(request.DaycareId),
		Min:        request.Min,
		Max:        request.Max,
		MinUnit:    store.DbNullString(request.MinUnit),
		MaxUnit:    store.DbNullString(request.MaxUnit),
		Stage:      store.DbNullString(request.Stage),
	}
}

func (c *AgeRangeService) AddAgeRange(ctx context.Context, request AgeRangeTransport) (store.AgeRange, error) {
	if claims.IsAdmin(ctx) && request.DaycareId == "" {
		return store.AgeRange{}, errors.New("as an admin, you must specify the a daycareId")
	} else {
		// default to requester daycare (e.g office manager)
		if request.DaycareId == "" {
			request.DaycareId = claims.GetDaycareId(ctx)
		}
	}

	ageRange, err := c.Store.AddAgeRange(nil, c.transportToStore(request))
	if err != nil {
		return store.AgeRange{}, errors.Wrap(err, "failed to add age range")
	}

	return ageRange, nil
}

func (c *AgeRangeService) GetAgeRange(ctx context.Context, request AgeRangeTransport) (store.AgeRange, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	ageRange, err := c.Store.GetAgeRange(nil, request.Id, searchOptions)
	if err != nil {
		return ageRange, errors.Wrap(err, "failed to get age range")
	}

	return ageRange, nil
}

func (c *AgeRangeService) DeleteAgeRange(ctx context.Context, request AgeRangeTransport) error {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	_, err := c.Store.GetAgeRange(nil, request.Id, searchOptions)
	if err != nil {
		return errors.Wrap(err, "failed to delete age range")
	}

	if err := c.Store.DeleteAgeRange(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete age range")
	}

	return nil
}

func (c *AgeRangeService) ListAgeRange(ctx context.Context) ([]store.AgeRange, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	ageRanges, err := c.Store.ListAgeRange(nil, searchOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list age ranges")
	}

	return ageRanges, nil
}

func (c *AgeRangeService) UpdateAgeRange(ctx context.Context, request AgeRangeTransport) (store.AgeRange, error) {
	if request.Id == "" {
		return store.AgeRange{}, ErrEmptyAgeRange
	}

	searchOptions := claims.GetDefaultSearchOptions(ctx)
	_, err := c.Store.GetAgeRange(nil, request.Id, searchOptions)
	if err != nil {
		return store.AgeRange{}, errors.Wrap(err, "failed to update age range")
	}

	ageRange, err := c.Store.UpdateAgeRange(nil, c.transportToStore(request))
	if err != nil {
		return ageRange, errors.Wrap(err, "failed to update age range")
	}

	return ageRange, nil
}

// ServiceMiddleware is a chainable behavior modifier for ageRangeService.
type ServiceMiddleware func(AgeRangeService) AgeRangeService
