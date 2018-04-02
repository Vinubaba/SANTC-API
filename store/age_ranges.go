package store

import (
	"database/sql"
	"errors"

	"github.com/jinzhu/gorm"
)

type AgeRange struct {
	AgeRangeId sql.NullString
	DaycareId  sql.NullString
	Stage      sql.NullString
	Min        int
	MinUnit    sql.NullString
	Max        int
	MaxUnit    sql.NullString
}

var (
	ErrAgeRangeNotFound = errors.New("age range not found")
)

func (s *Store) AddAgeRange(tx *gorm.DB, ageRange AgeRange) (AgeRange, error) {
	db := s.dbOrTx(tx)

	ageRange.AgeRangeId = DbNullString(s.StringGenerator.GenerateUuid())

	if err := db.Create(&ageRange).Error; err != nil {
		return AgeRange{}, err
	}

	return ageRange, nil
}

func (s *Store) DeleteAgeRange(tx *gorm.DB, ageRangeId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.ageRangeExists(db, ageRangeId) {
		return ErrAgeRangeNotFound
	}

	if err := db.Where("age_range_id = ?", ageRangeId).Delete(&AgeRange{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) ageRangeExists(tx *gorm.DB, ageRangeId string) bool {
	c := AgeRange{AgeRangeId: sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}}
	return !tx.Model(AgeRange{}).Where("age_range_id = ?", ageRangeId).First(&c).RecordNotFound()
}

func (s *Store) GetAgeRange(tx *gorm.DB, ageRangeId string, options SearchOptions) (AgeRange, error) {
	db := s.dbOrTx(tx)

	query := db.Table("age_ranges").
		Select("age_ranges.age_range_id," +
			"age_ranges.daycare_id," +
			"age_ranges.stage," +
			"age_ranges.min," +
			"age_ranges.min_unit," +
			"age_ranges.max," +
			"age_ranges.max_unit")
	if options.DaycareId != "" {
		query = query.Where("age_ranges.daycare_id = ?", options.DaycareId)
	}
	query = query.Where("age_ranges.age_range_id = ?", ageRangeId)

	rows, err := query.Rows()
	if err != nil {
		return AgeRange{}, err
	}

	ageRanges, err := s.scanAgeRangeRows(rows)
	if err != nil {
		return AgeRange{}, err
	}
	if len(ageRanges) == 0 {
		return AgeRange{}, ErrAgeRangeNotFound
	}

	return ageRanges[0], nil
}

func (s *Store) scanAgeRangeRows(rows *sql.Rows) ([]AgeRange, error) {
	ageRanges := []AgeRange{}
	for rows.Next() {
		currentAgeRange := AgeRange{}
		if err := rows.Scan(&currentAgeRange.AgeRangeId,
			&currentAgeRange.DaycareId,
			&currentAgeRange.Stage,
			&currentAgeRange.Min,
			&currentAgeRange.MinUnit,
			&currentAgeRange.Max,
			&currentAgeRange.MaxUnit,
		); err != nil {
			return []AgeRange{}, err
		}
		ageRanges = append(ageRanges, currentAgeRange)
	}

	return ageRanges, nil
}

func (s *Store) ListAgeRange(tx *gorm.DB, options SearchOptions) ([]AgeRange, error) {
	db := s.dbOrTx(tx)

	query := db.Table("age_ranges").
		Select("age_ranges.age_range_id," +
			"age_ranges.daycare_id," +
			"age_ranges.stage," +
			"age_ranges.min," +
			"age_ranges.min_unit," +
			"age_ranges.max," +
			"age_ranges.max_unit")
	if options.DaycareId != "" {
		query = query.Where("age_ranges.daycare_id = ?", options.DaycareId)
	}

	rows, err := query.Rows()
	if err != nil {
		return []AgeRange{}, err
	}

	ageRanges, err := s.scanAgeRangeRows(rows)
	if err != nil {
		return []AgeRange{}, err
	}

	return ageRanges, nil
}

func (s *Store) UpdateAgeRange(tx *gorm.DB, ageRange AgeRange) (AgeRange, error) {
	db := s.dbOrTx(tx)

	res := db.Where("age_range_id = ?", ageRange.AgeRangeId).Model(&AgeRange{}).Updates(ageRange).First(&ageRange)
	if res.RecordNotFound() {
		return AgeRange{}, ErrAgeRangeNotFound
	}
	if err := res.Error; err != nil {
		return AgeRange{}, err
	}

	return s.GetAgeRange(db, ageRange.AgeRangeId.String, SearchOptions{})
}
