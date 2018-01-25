package store

import "context"

type Allergy struct {
	AllergyId string
	ChildId   string
	Allergy   string
}

func (s *Store) AddAllergy(ctx context.Context, allergy Allergy) (Allergy, error) {
	allergy.AllergyId = s.StringGenerator.GenerateUuid()
	if err := s.Db.Create(&allergy).Error; err != nil {
		return Allergy{}, err
	}

	return allergy, nil
}

func (s *Store) DeleteAllergy(ctx context.Context, allergyId string) error {
	if err := s.Db.Delete(&Allergy{AllergyId: allergyId}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) FindAllergiesOfChild(ctx context.Context, childId string) ([]Allergy, error) {
	allergies := []Allergy{}
	if err := s.Db.Where("child_id = ?", childId).Find(&allergies).Error; err != nil {
		return nil, err
	}

	return allergies, nil
}

func (s *Store) RemoveAllergiesOfChild(ctx context.Context, childId string) error {
	return s.Db.Where("child_id = ?", childId).Delete(&Allergy{}).Error
}
