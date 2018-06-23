package schedules

import (
	"context"
	"regexp"

	. "github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/firebase/claims"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrEmptySchedule    = errors.New("scheduleId cannot be empty")
	ErrBadTimeFormat    = errors.New("time does not match the following regex: " + timeRegexp.String())
	ErrDifferentDaycare = errors.New("")
)

type Service interface {
	AddSchedule(ctx context.Context, request ScheduleTransport) (store.Schedule, error)
	DeleteSchedule(ctx context.Context, request ScheduleTransport) error
	UpdateSchedule(ctx context.Context, request ScheduleTransport) (store.Schedule, error)
	GetSchedule(ctx context.Context, request ScheduleTransport) (store.Schedule, error)
	ListSchedules(ctx context.Context) ([]store.Schedule, error)
}

type ScheduleService struct {
	Store interface {
		Tx() *gorm.DB

		AddSchedule(tx *gorm.DB, schedule store.Schedule) (store.Schedule, error)
		UpdateSchedule(tx *gorm.DB, schedule store.Schedule) (store.Schedule, error)
		GetSchedule(tx *gorm.DB, scheduleId string, options store.SearchOptions) (store.Schedule, error)
		ListSchedules(tx *gorm.DB, options store.SearchOptions) ([]store.Schedule, error)
		DeleteSchedule(tx *gorm.DB, scheduleId string) error

		GetChild(tx *gorm.DB, childId string, options store.SearchOptions) (store.Child, error)
		GetUser(tx *gorm.DB, userId string, searchOptions store.SearchOptions) (store.User, error)
		UpdateChild(tx *gorm.DB, child store.Child) error
		UpdateUser(tx *gorm.DB, user store.User) (store.User, error)
	} `inject:""`
	Logger *log.Logger `inject:""`
}

func (c *ScheduleService) transportToStore(request ScheduleTransport) store.Schedule {
	return store.Schedule{
		ScheduleId:     store.DbNullString(request.Id),
		WalkIn:         store.DbNullBool(request.WalkIn),
		MondayStart:    store.DbNullString(request.MondayStart),
		MondayEnd:      store.DbNullString(request.MondayEnd),
		TuesdayStart:   store.DbNullString(request.TuesdayStart),
		TuesdayEnd:     store.DbNullString(request.TuesdayEnd),
		WednesdayStart: store.DbNullString(request.WednesdayStart),
		WednesdayEnd:   store.DbNullString(request.WednesdayEnd),
		ThursdayStart:  store.DbNullString(request.ThursdayStart),
		ThursdayEnd:    store.DbNullString(request.ThursdayEnd),
		FridayStart:    store.DbNullString(request.FridayStart),
		FridayEnd:      store.DbNullString(request.FridayEnd),
		SaturdayStart:  store.DbNullString(request.SaturdayStart),
		SaturdayEnd:    store.DbNullString(request.SaturdayEnd),
		SundayStart:    store.DbNullString(request.SundayStart),
		SundayEnd:      store.DbNullString(request.SundayEnd),
	}
}

var (
	timeRegexp = regexp.MustCompile(`^\d{1,2}:\d{2}\s(AM|PM)$`)
)

func (c *ScheduleService) validateRequest(request ScheduleTransport) error {
	toValidate := make([]*string, 0)
	toValidate = append(toValidate,
		request.MondayStart,
		request.MondayEnd,
		request.TuesdayStart,
		request.TuesdayEnd,
		request.WednesdayStart,
		request.WednesdayEnd,
		request.TuesdayStart,
		request.TuesdayEnd,
		request.FridayStart,
		request.FridayEnd,
		request.SaturdayStart,
		request.SaturdayEnd,
		request.SundayStart,
		request.SundayEnd)

	for _, value := range toValidate {
		if value != nil {
			if !timeRegexp.MatchString(*request.MondayStart) {
				return errors.Wrap(ErrBadTimeFormat, *value+" does not match regex")
			}
		}
	}

	return nil
}

func (c *ScheduleService) AddSchedule(ctx context.Context, request ScheduleTransport) (store.Schedule, error) {
	if err := c.validateRequest(request); err != nil {
		return store.Schedule{}, errors.Wrap(err, "failed to validate request")
	}

	if request.ChildId != nil {
		_, err := c.Store.GetChild(nil, *request.ChildId, claims.GetDefaultSearchOptions(ctx))
		if err != nil {
			return store.Schedule{}, errors.Wrap(err, "failed to get child")
		}
	}
	if request.TeacherId != nil {
		_, err := c.Store.GetUser(nil, *request.TeacherId, claims.GetDefaultSearchOptions(ctx))
		if err != nil {
			return store.Schedule{}, errors.Wrap(err, "failed to get teacher")
		}
	}

	transaction := c.Store.Tx()
	defer transaction.Close()

	schedule, err := c.Store.AddSchedule(transaction, c.transportToStore(request))
	if err != nil {
		transaction.Rollback()
		return store.Schedule{}, errors.Wrap(err, "failed to add schedule")
	}

	if request.ChildId != nil {
		if err := c.Store.UpdateChild(transaction, store.Child{
			ChildId:    store.DbNullString(request.ChildId),
			ScheduleId: schedule.ScheduleId,
		}); err != nil {
			transaction.Rollback()
			return store.Schedule{}, errors.Wrap(err, "failed to update child with scheduleId")
		}
	}
	if request.TeacherId != nil {
		if _, err := c.Store.UpdateUser(transaction, store.User{
			UserId:     store.DbNullString(request.TeacherId),
			ScheduleId: schedule.ScheduleId,
		}); err != nil {
			transaction.Rollback()
			return store.Schedule{}, errors.Wrap(err, "failed to update teacher with scheduleId")
		}
	}
	transaction.Commit()

	return schedule, nil
}

func (c *ScheduleService) GetSchedule(ctx context.Context, request ScheduleTransport) (store.Schedule, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	if !IsNilOrEmpty(request.TeacherId) {
		_, err := c.Store.GetUser(nil, *request.TeacherId, searchOptions)
		if err != nil {
			return store.Schedule{}, errors.Wrap(err, "failed to get schedule")
		}
		searchOptions.TeacherId = *request.TeacherId
	}
	if !IsNilOrEmpty(request.ChildId) {
		_, err := c.Store.GetChild(nil, *request.ChildId, searchOptions)
		if err != nil {
			return store.Schedule{}, errors.Wrap(err, "failed to get schedule")
		}
		searchOptions.ChildrenId = append(searchOptions.ChildrenId, *request.ChildId)
	}

	schedule, err := c.Store.GetSchedule(nil, *request.Id, searchOptions)
	if err != nil {
		return schedule, errors.Wrap(err, "failed to get schedule")
	}

	return schedule, nil
}

func (c *ScheduleService) DeleteSchedule(ctx context.Context, request ScheduleTransport) error {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	if !IsNilOrEmpty(request.TeacherId) {
		_, err := c.Store.GetUser(nil, *request.TeacherId, searchOptions)
		if err != nil {
			return errors.Wrap(err, "failed to delete schedule")
		}
		searchOptions.TeacherId = *request.TeacherId
	}
	if !IsNilOrEmpty(request.ChildId) {
		_, err := c.Store.GetChild(nil, *request.ChildId, searchOptions)
		if err != nil {
			return errors.Wrap(err, "failed to delete schedule")
		}
		searchOptions.ChildrenId = append(searchOptions.ChildrenId, *request.ChildId)
	}
	_, err := c.Store.GetSchedule(nil, *request.Id, searchOptions)
	if err != nil {
		return errors.Wrap(err, "failed to delete schedule")
	}

	if err := c.Store.DeleteSchedule(nil, *request.Id); err != nil {
		return errors.Wrap(err, "failed to delete schedule")
	}

	return nil
}

func (c *ScheduleService) ListSchedules(ctx context.Context) ([]store.Schedule, error) {
	schedules, err := c.Store.ListSchedules(nil, claims.GetDefaultSearchOptions(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to list schedules")
	}

	return schedules, nil
}

func (c *ScheduleService) UpdateSchedule(ctx context.Context, request ScheduleTransport) (store.Schedule, error) {
	if err := c.validateRequest(request); err != nil {
		return store.Schedule{}, errors.Wrap(err, "failed to validate request")
	}

	if request.Id == nil {
		return store.Schedule{}, ErrEmptySchedule
	}

	searchOptions := claims.GetDefaultSearchOptions(ctx)
	if request.TeacherId != nil {
		_, err := c.Store.GetUser(nil, *request.TeacherId, searchOptions)
		if err != nil {
			return store.Schedule{}, errors.Wrap(err, "failed to get teacher")
		}

		searchOptions.TeacherId = *request.TeacherId
	}
	if request.ChildId != nil {
		_, err := c.Store.GetChild(nil, *request.ChildId, searchOptions)
		if err != nil {
			return store.Schedule{}, errors.Wrap(err, "failed to get child")
		}
		searchOptions.ChildrenId = append(searchOptions.ChildrenId, *request.ChildId)
	}

	_, err := c.Store.GetSchedule(nil, *request.Id, searchOptions)
	if err != nil {
		return store.Schedule{}, errors.Wrap(err, "failed to update schedule")
	}

	schedule, err := c.Store.UpdateSchedule(nil, c.transportToStore(request))
	if err != nil {
		return schedule, errors.Wrap(err, "failed to update schedule")
	}

	return schedule, nil
}

// ServiceMiddleware is a chainable behavior modifier for scheduleService.
type ServiceMiddleware func(ScheduleService) ScheduleService
