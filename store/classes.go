package store

import (
	"database/sql"
	"errors"

	"github.com/jinzhu/gorm"
)

var (
	ErrClassNotFound = errors.New("class not found")
)

type Class struct {
	ClassId     sql.NullString
	Name        sql.NullString
	Description sql.NullString
	ImageUri    sql.NullString
}

func (s *Store) AddClass(tx *gorm.DB, class Class) (Class, error) {
	db := s.dbOrTx(tx)

	class.ClassId = DbNullString(s.StringGenerator.GenerateUuid())

	if err := db.Create(&class).Error; err != nil {
		return Class{}, err
	}

	return class, nil
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

	rows, err := db.Table("classes").
		Raw("SELECT classes.class_id,"+
			"classes.name,"+
			"classes.description,"+
			"classes.image_uri"+
			" FROM classes"+
			" WHERE classes.class_id = ?", classId).
		Rows()
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

func (s *Store) scanClassRows(rows *sql.Rows) ([]Class, error) {
	classes := []Class{}
	for rows.Next() {
		currentClass := Class{}
		if err := rows.Scan(&currentClass.ClassId,
			&currentClass.Name,
			&currentClass.Description,
			&currentClass.ImageUri); err != nil {
			return []Class{}, err
		}
		classes = append(classes, currentClass)
	}

	return classes, nil
}

func (s *Store) ListClass(tx *gorm.DB) ([]Class, error) {
	db := s.dbOrTx(tx)

	rows, err := db.Table("classes").
		Raw("SELECT classes.class_id," +
			"classes.name," +
			"classes.description," +
			"classes.image_uri" +
			" FROM classes").
		Rows()

	if err != nil {
		return []Class{}, err
	}
	classes, err := s.scanClassRows(rows)
	if err != nil {
		return []Class{}, err
	}

	if len(classes) == 0 {
		return []Class{}, ErrClassNotFound
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
		return Class{}, err
	}

	return s.GetClass(db, class.ClassId.String)
}
