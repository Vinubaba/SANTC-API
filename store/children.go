package store

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrChildNotFound = errors.New("child not found")
)

type Child struct {
	ChildId     string
	FirstName   string
	LastName    string
	BirthDate   time.Time
	Gender      string
	PicturePath string
}

func (s *Store) AddChild(tx *gorm.DB, child Child) (Child, error) {
	db := s.dbOrTx(tx)

	child.ChildId = s.StringGenerator.GenerateUuid()

	if err := db.Create(&child).Error; err != nil {
		return Child{}, err
	}

	return child, nil
}

func (s *Store) DeleteChild(tx *gorm.DB, childId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.childExists(db, childId) {
		return ErrChildNotFound
	}

	if err := db.Where("child_id = ?", childId).Delete(&Child{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) childExists(tx *gorm.DB, childId string) bool {
	c := Child{ChildId: childId}
	return !tx.Model(Child{}).Where("child_id = ?", childId).First(&c).RecordNotFound()
}

func (s *Store) GetChild(tx *gorm.DB, childId string) (Child, error) {
	db := s.dbOrTx(tx)

	child := Child{}
	res := db.Where("child_id = ?", childId).First(&child)
	if res.RecordNotFound() {
		return Child{}, ErrChildNotFound
	}
	if err := res.Error; err != nil {
		return Child{}, err
	}

	return child, nil
}

func (s *Store) ListChild(tx *gorm.DB) ([]Child, error) {
	db := s.dbOrTx(tx)

	children := []Child{}
	if err := db.Find(&children).Error; err != nil {
		return nil, err
	}

	return children, nil
}

func (s *Store) UpdateChild(tx *gorm.DB, child Child) (Child, error) {
	db := s.dbOrTx(tx)

	res := db.Where("child_id = ?", child.ChildId).Model(&Child{}).Updates(child).First(&child)
	if res.RecordNotFound() {
		return Child{}, ErrChildNotFound
	}
	if err := res.Error; err != nil {
		return Child{}, err
	}

	return child, nil
}
