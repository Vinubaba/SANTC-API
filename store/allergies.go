package store

import (
	"errors"
	"github.com/jinzhu/gorm"
	"strings"
)

type Allergies []Allergy

func (s *Allergies) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case string:
		instructions := strings.Split(v, ",")
		for _, instruction := range instructions {
			*s = append(*s, Allergy{Allergy: instruction})
		}
	default:
		return errors.New("need string with roles separated by virgula")
	}
	return nil
}

func (s Allergies) ToList() []string {
	allergies := make([]string, 0)
	for _, allergy := range s {
		allergies = append(allergies, allergy.Allergy)
	}
	return allergies
}

type Allergy struct {
	AllergyId string
	ChildId   string
	Allergy   string
}

func (s *Store) AddAllergy(tx *gorm.DB, allergy Allergy) (Allergy, error) {
	db := s.dbOrTx(tx)

	allergy.AllergyId = s.StringGenerator.GenerateUuid()
	if err := db.Create(&allergy).Error; err != nil {
		return Allergy{}, err
	}

	return allergy, nil
}

func (s *Store) DeleteAllergy(tx *gorm.DB, allergyId string) error {
	db := s.dbOrTx(tx)

	if err := db.Delete(&Allergy{AllergyId: allergyId}).Error; err != nil {
		return err
	}

	return nil
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
