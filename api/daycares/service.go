package daycares

import (
	"context"

	. "github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/firebase/claims"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrEmptyDaycare = errors.New("daycareId cannot be empty")
)

type Service interface {
	AddDaycare(ctx context.Context, request DaycareTransport) (store.Daycare, error)
	DeleteDaycare(ctx context.Context, request DaycareTransport) error
	UpdateDaycare(ctx context.Context, request DaycareTransport) (store.Daycare, error)
	GetDaycare(ctx context.Context, request DaycareTransport) (store.Daycare, error)
	ListDaycare(ctx context.Context) ([]store.Daycare, error)
}

type DaycareService struct {
	Store interface {
		Tx() *gorm.DB

		AddDaycare(tx *gorm.DB, daycare store.Daycare) (store.Daycare, error)
		UpdateDaycare(tx *gorm.DB, daycare store.Daycare) (store.Daycare, error)
		GetDaycare(tx *gorm.DB, daycareId string, options store.SearchOptions) (store.Daycare, error)
		ListDaycare(tx *gorm.DB, options store.SearchOptions) ([]store.Daycare, error)
		DeleteDaycare(tx *gorm.DB, daycareId string) error
	} `inject:""`
	Logger *log.Logger `inject:""`
}

func (c *DaycareService) transportToStore(request DaycareTransport) store.Daycare {
	return store.Daycare{
		DaycareId: store.DbNullString(request.Id),
		Name:      store.DbNullString(request.Name),
		Zip:       store.DbNullString(request.Zip),
		State:     store.DbNullString(request.State),
		City:      store.DbNullString(request.City),
		Address_1: store.DbNullString(request.Address_1),
		Address_2: store.DbNullString(request.Address_2),
	}
}

func (c *DaycareService) AddDaycare(ctx context.Context, request DaycareTransport) (store.Daycare, error) {
	daycare, err := c.Store.AddDaycare(nil, c.transportToStore(request))
	if err != nil {
		return store.Daycare{}, errors.Wrap(err, "failed to add daycare")
	}

	return daycare, nil
}

func (c *DaycareService) GetDaycare(ctx context.Context, request DaycareTransport) (store.Daycare, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	daycare, err := c.Store.GetDaycare(nil, *request.Id, searchOptions)
	if err != nil {
		return daycare, errors.Wrap(err, "failed to get daycare")
	}

	return daycare, nil
}

func (c *DaycareService) DeleteDaycare(ctx context.Context, request DaycareTransport) error {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	_, err := c.Store.GetDaycare(nil, *request.Id, searchOptions)
	if err != nil {
		return errors.Wrap(err, "failed to delete daycare")
	}

	if err := c.Store.DeleteDaycare(nil, *request.Id); err != nil {
		return errors.Wrap(err, "failed to delete daycare")
	}

	return nil
}

func (c *DaycareService) ListDaycare(ctx context.Context) ([]store.Daycare, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	daycares, err := c.Store.ListDaycare(nil, searchOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list daycares")
	}

	return daycares, nil
}

func (c *DaycareService) UpdateDaycare(ctx context.Context, request DaycareTransport) (store.Daycare, error) {
	if IsNilOrEmpty(request.Id) {
		return store.Daycare{}, ErrEmptyDaycare
	}

	searchOptions := claims.GetDefaultSearchOptions(ctx)
	_, err := c.Store.GetDaycare(nil, *request.Id, searchOptions)
	if err != nil {
		return store.Daycare{}, errors.Wrap(err, "failed to update daycare")
	}

	daycare, err := c.Store.UpdateDaycare(nil, c.transportToStore(request))
	if err != nil {
		return daycare, errors.Wrap(err, "failed to update daycare")
	}

	return daycare, nil
}

// ServiceMiddleware is a chainable behavior modifier for daycareService.
type ServiceMiddleware func(DaycareService) DaycareService
