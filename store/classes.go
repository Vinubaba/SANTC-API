package store

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
)

var (
	ErrClassNotFound          = errors.New("class not found")
	ErrClassNameAlreadyExists = errors.New("class name already exists")
)

type Class struct {
	ClassId     sql.NullString
	DaycareId   sql.NullString
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

	return s.GetClass(db, class.ClassId.String, SearchOptions{})
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

func (s *Store) GetClass(tx *gorm.DB, classId string, options SearchOptions) (Class, error) {
	db := s.dbOrTx(tx)
	query := db.Table("classes").
		Select("classes.class_id," +
			"classes.daycare_id," +
			"classes.age_range_id," +
			"classes.name," +
			"classes.description," +
			"classes.image_uri," +
			"age_ranges.age_range_id," +
			"age_ranges.daycare_id," +
			"age_ranges.stage," +
			"age_ranges.min," +
			"age_ranges.min_unit," +
			"age_ranges.max," +
			"age_ranges.max_unit").
		Joins("left join age_ranges ON age_ranges.age_range_id = classes.age_range_id")
	if options.DaycareId != "" {
		query = query.Where("classes.daycare_id = ?", options.DaycareId)
	}
	query = query.Where("classes.class_id = ?", classId)

	rows, err := query.Rows()
	if err != nil {
		return Class{}, err
	}
	classes, err := s.scanClassRows(rows)
	if err != nil {
		return Class{}, err
	}

	if len(classes) == 0 {
		return Class{}, ErrClassNotFound
	}

	return classes[0], nil
}

func (s *Store) ListClasses(tx *gorm.DB, options SearchOptions) ([]Class, error) {
	db := s.dbOrTx(tx)

	query := db.Table("classes").
		Select("classes.class_id," +
			"classes.daycare_id," +
			"classes.age_range_id," +
			"classes.name," +
			"classes.description," +
			"classes.image_uri," +
			"age_ranges.age_range_id," +
			"age_ranges.daycare_id," +
			"age_ranges.stage," +
			"age_ranges.min," +
			"age_ranges.min_unit," +
			"age_ranges.max," +
			"age_ranges.max_unit").
		Joins("left join age_ranges ON age_ranges.age_range_id = classes.age_range_id")
	if options.DaycareId != "" {
		query = query.Where("classes.daycare_id = ?", options.DaycareId)
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}

	return s.scanClassRows(rows)
}

func (s *Store) scanClassRows(rows *sql.Rows) ([]Class, error) {
	classes := []Class{}
	for rows.Next() {
		currentClass := Class{}
		if err := rows.Scan(&currentClass.ClassId,
			&currentClass.DaycareId,
			&currentClass.AgeRangeId,
			&currentClass.Name,
			&currentClass.Description,
			&currentClass.ImageUri,
			&currentClass.AgeRange.AgeRangeId,
			&currentClass.AgeRange.DaycareId,
			&currentClass.AgeRange.Stage,
			&currentClass.AgeRange.Min,
			&currentClass.AgeRange.MinUnit,
			&currentClass.AgeRange.Max,
			&currentClass.AgeRange.MaxUnit,
		); err != nil {
			return []Class{}, err
		}
		classes = append(classes, currentClass)
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

	return s.GetClass(db, class.ClassId.String, SearchOptions{})
}
