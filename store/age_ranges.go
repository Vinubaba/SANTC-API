package store

import (
	"database/sql"
	"errors"

	"github.com/jinzhu/gorm"
)

type AgeRange struct {
	AgeRangeId sql.NullString
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

func (s *Store) GetAgeRange(tx *gorm.DB, ageRangeId string) (AgeRange, error) {
	db := s.dbOrTx(tx)

	ageRange := AgeRange{
		AgeRangeId: DbNullString(ageRangeId),
	}
	if err := db.Model(Class{}).Where("age_range_id = ?", ageRangeId).First(&ageRange).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return AgeRange{}, ErrAgeRangeNotFound
		}
		return AgeRange{}, err
	}
	return ageRange, nil
}

func (s *Store) ListAgeRange(tx *gorm.DB) ([]AgeRange, error) {
	db := s.dbOrTx(tx)

	ageRanges := make([]AgeRange, 0)
	if err := db.Model(AgeRange{}).Find(&ageRanges).Error; err != nil {
		return nil, err
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

	return s.GetAgeRange(db, ageRange.AgeRangeId.String)
}
