package store

import (
	"database/sql"
	"errors"

	"github.com/jinzhu/gorm"
)

var (
	ErrAllergyNotFound = errors.New("allergy not found")
)

type Allergy struct {
	AllergyId   sql.NullString
	ChildId     sql.NullString
	Allergy     sql.NullString
	Instruction sql.NullString
}

type Allergies []Allergy

func (a *Allergies) add(allergy Allergy) {
	if allergy.AllergyId.String == "" {
		return
	}
	for _, al := range *a {
		if al.AllergyId.String == allergy.AllergyId.String {
			return
		}
	}
	*a = append(*a, allergy)
}

func (s *Store) AddAllergy(tx *gorm.DB, allergy Allergy) (Allergy, error) {
	db := s.dbOrTx(tx)
	allergy.AllergyId = s.newId()
	if err := db.Create(&allergy).Error; err != nil {
		return Allergy{}, err
	}

	return allergy, nil
}

func (s *Store) DeleteAllergy(tx *gorm.DB, allergyId string) error {
	db := s.dbOrTx(tx)

	if !s.allergyExists(db, allergyId) {
		return ErrAllergyNotFound
	}

	return db.Where("allergy_id = ?", allergyId).Delete(&Class{}).Error
}

func (s *Store) allergyExists(tx *gorm.DB, allergyId string) bool {
	c := Allergy{AllergyId: sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}}
	return !tx.Model(Class{}).Where("allergy_id = ?", allergyId).First(&c).RecordNotFound()
}

func (s *Store) FindAllergiesOfChild(tx *gorm.DB, childId string) ([]Allergy, error) {
	db := s.dbOrTx(tx)

	allergies := []Allergy{}
	if err := db.Where("child_id = ?", childId).Find(&allergies).Error; err != nil {
		return nil, err
	}

	return allergies, nil
}

func (s *Store) RemoveAllergiesOfChild(tx *gorm.DB, childId string) error {
	db := s.dbOrTx(tx)
	return db.Where("child_id = ?", childId).Delete(&Allergy{}).Error
}
