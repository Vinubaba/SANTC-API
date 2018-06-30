package store

import (
	"database/sql"
	"errors"

	"github.com/jinzhu/gorm"
)

type Schedule struct {
	ScheduleId     sql.NullString
	WalkIn         sql.NullBool
	MondayStart    sql.NullString
	MondayEnd      sql.NullString
	TuesdayStart   sql.NullString
	TuesdayEnd     sql.NullString
	WednesdayStart sql.NullString
	WednesdayEnd   sql.NullString
	ThursdayStart  sql.NullString
	ThursdayEnd    sql.NullString
	FridayStart    sql.NullString
	FridayEnd      sql.NullString
	SaturdayStart  sql.NullString
	SaturdayEnd    sql.NullString
	SundayStart    sql.NullString
	SundayEnd      sql.NullString
}

var (
	ErrScheduleNotFound = errors.New("schedule not found")
)

func (s *Store) AddSchedule(tx *gorm.DB, schedule Schedule) (Schedule, error) {
	db := s.dbOrTx(tx)

	schedule.ScheduleId = s.newId()

	if err := db.Create(&schedule).Error; err != nil {
		return Schedule{}, err
	}

	return schedule, nil
}

func (s *Store) DeleteSchedule(tx *gorm.DB, scheduleId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.scheduleExists(db, scheduleId) {
		return ErrScheduleNotFound
	}

	if err := db.Where("schedule_id = ?", scheduleId).Delete(&Schedule{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) scheduleExists(tx *gorm.DB, scheduleId string) bool {
	c := Schedule{ScheduleId: sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}}
	return !tx.Model(Schedule{}).Where("schedule_id = ?", scheduleId).First(&c).RecordNotFound()
}

func (s *Store) baseQuery(tx *gorm.DB, options SearchOptions) *gorm.DB {
	db := s.dbOrTx(tx)
	query := db.Table("schedules, users, children").
		Select("schedules.schedule_id," +
			"schedules.walk_in," +
			"schedules.monday_start," +
			"schedules.monday_end," +
			"schedules.tuesday_start," +
			"schedules.tuesday_end," +
			"schedules.wednesday_start," +
			"schedules.wednesday_end," +
			"schedules.thursday_start," +
			"schedules.thursday_end," +
			"schedules.friday_start," +
			"schedules.friday_end," +
			"schedules.saturday_start," +
			"schedules.saturday_end," +
			"schedules.sunday_start," +
			"schedules.sunday_end")

	if options.TeacherId != "" {
		query = query.Where("users.user_id = ?", options.TeacherId)
		query = query.Where("users.schedule_id = schedules.schedule_id")
	}
	if len(options.ChildrenId) > 0 {
		query = query.Where("children.child_id IN (?) AND children.schedule_id = schedules.schedule_id", options.ChildrenId)
	}

	return query
}

func (s *Store) GetSchedule(tx *gorm.DB, scheduleId string, options SearchOptions) (Schedule, error) {
	db := s.dbOrTx(tx)

	query := s.baseQuery(db, options)
	query = query.Where("schedules.schedule_id = ?", scheduleId)

	rows, err := query.Rows()
	if err != nil {
		return Schedule{}, err
	}

	schedules, err := s.scanScheduleRows(rows)
	if err != nil {
		return Schedule{}, err
	}
	if len(schedules) == 0 {
		return Schedule{}, ErrScheduleNotFound
	}

	return schedules[0], nil
}

func (s *Store) scanScheduleRows(rows *sql.Rows) ([]Schedule, error) {
	schedules := []Schedule{}
	for rows.Next() {
		currentSchedule := Schedule{}
		if err := rows.Scan(&currentSchedule.ScheduleId,
			&currentSchedule.WalkIn,
			&currentSchedule.MondayStart,
			&currentSchedule.MondayEnd,
			&currentSchedule.TuesdayStart,
			&currentSchedule.TuesdayEnd,
			&currentSchedule.WednesdayStart,
			&currentSchedule.WednesdayEnd,
			&currentSchedule.ThursdayStart,
			&currentSchedule.ThursdayEnd,
			&currentSchedule.FridayStart,
			&currentSchedule.FridayEnd,
			&currentSchedule.SaturdayStart,
			&currentSchedule.SaturdayEnd,
			&currentSchedule.SundayStart,
			&currentSchedule.SundayEnd,
		); err != nil {
			return []Schedule{}, err
		}
		schedules = append(schedules, currentSchedule)
	}

	return schedules, nil
}

func (s *Store) ListSchedules(tx *gorm.DB, options SearchOptions) ([]Schedule, error) {
	db := s.dbOrTx(tx)

	query := s.baseQuery(db, options)

	rows, err := query.Rows()
	if err != nil {
		return []Schedule{}, err
	}

	schedules, err := s.scanScheduleRows(rows)
	if err != nil {
		return []Schedule{}, err
	}

	return schedules, nil
}

func (s *Store) UpdateSchedule(tx *gorm.DB, schedule Schedule) (Schedule, error) {
	db := s.dbOrTx(tx)

	// Ensure we don't update the id
	id := schedule.ScheduleId
	schedule.ScheduleId.String = ""
	schedule.ScheduleId.Valid = false

	res := db.Where("schedule_id = ?", id).Model(&Schedule{}).Updates(schedule).First(&schedule)
	if res.RecordNotFound() {
		return Schedule{}, ErrScheduleNotFound
	}
	if err := res.Error; err != nil {
		return Schedule{}, err
	}

	return s.GetSchedule(db, schedule.ScheduleId.String, SearchOptions{})
}
