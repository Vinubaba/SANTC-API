package store

import (
	"database/sql"
	"errors"

	"github.com/jinzhu/gorm"
	"strings"
)

var (
	ErrClassNotFound          = errors.New("class not found")
	ErrClassNameAlreadyExists = errors.New("class name already exists")
)

type Class struct {
	ClassId     sql.NullString
	AgeRangeId  sql.NullString
	Name        sql.NullString
	Description sql.NullString
	ImageUri    sql.NullString
	AgeRange    AgeRange `sql:"-" gorm:"foreignkey:AgeRangeId association_foreignkey:AgeRangeId"`
}

func (s *Store) AddClass(tx *gorm.DB, class Class) (Class, error) {
	db := s.dbOrTx(tx)

	if class.AgeRange.AgeRangeId.String == "" {
		ageRange, err := s.AddAgeRange(db, class.AgeRange)
		if err != nil {
			return Class{}, err
		}
		class.AgeRangeId = ageRange.AgeRangeId
	}

	class.ClassId = DbNullString(s.StringGenerator.GenerateUuid())
	if err := db.Create(&class).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"classes_name_key\"") {
			return Class{}, ErrClassNameAlreadyExists
		}
		return Class{}, err
	}

	return s.GetClass(db, class.ClassId.String)
}

func (s *Store) DeleteClass(tx *gorm.DB, classId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.classExists(db, classId) {
		return ErrClassNotFound
	}

	if err := db.Where("class_id = ?", classId).Delete(&Class{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) classExists(tx *gorm.DB, classId string) bool {
	c := Class{ClassId: sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}}
	return !tx.Model(Class{}).Where("class_id = ?", classId).First(&c).RecordNotFound()
}

func (s *Store) GetClass(tx *gorm.DB, classId string) (Class, error) {
	db := s.dbOrTx(tx)

	class := Class{
		ClassId: DbNullString(classId),
	}
	if err := db.Model(Class{}).Where("class_id = ?", classId).First(&class).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Class{}, ErrClassNotFound
		}
		return Class{}, err
	}
	if err := db.Model(AgeRange{}).Where("age_range_id = ?", class.AgeRangeId).First(&class.AgeRange).Error; err != nil {
		return Class{}, err
	}

	return class, nil
}

func (s *Store) ListClass(tx *gorm.DB) ([]Class, error) {
	db := s.dbOrTx(tx)

	classes := make([]Class, 0)

	if err := db.Model(Class{}).Find(&classes).Error; err != nil {
		return nil, err
	}
	for i, class := range classes {
		if err := db.Model(classes[i]).Where("age_range_id = ?", class.AgeRangeId).First(&classes[i].AgeRange).Error; err != nil {
			return nil, err
		}
	}

	return classes, nil
}

func (s *Store) UpdateClass(tx *gorm.DB, class Class) (Class, error) {
	db := s.dbOrTx(tx)

	res := db.Where("class_id = ?", class.ClassId).Model(&Class{}).Updates(class).First(&class)
	if res.RecordNotFound() {
		return Class{}, ErrClassNotFound
	}
	if err := res.Error; err != nil {
		if strings.Contains(err.Error(), "insert or update on table \"classes\" violates foreign key constraint \"classes_age_range_id_fkey\"") {
			return Class{}, ErrAgeRangeNotFound
		}
		return Class{}, err
	}

	return s.GetClass(db, class.ClassId.String)
}
