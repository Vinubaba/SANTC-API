package store

import (
	"github.com/jinzhu/gorm"
)

type AdultResponsible struct {
	ResponsibleId string
	Email         string
	FirstName     string
	LastName      string
	Gender        string
	Phone         string
	Addres_1      string
	Addres_2      string
	City          string
	State         string
	Zip           string
}

func (s *Store) AddAdultResponsible(tx *gorm.DB, adult AdultResponsible) (AdultResponsible, error) {
	db := s.dbOrTx(tx)

	if err := db.Create(&adult).Error; err != nil {
		return AdultResponsible{}, err
	}

	return adult, nil
}

func (s *Store) ListAdultResponsible(tx *gorm.DB) ([]AdultResponsible, error) {
	db := s.dbOrTx(tx)

	adults := []AdultResponsible{}
	if err := db.Find(&adults).Error; err != nil {
		return nil, err
	}

	return adults, nil
}

func (s *Store) DeleteAdultResponsible(tx *gorm.DB, adultId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.userExists(db, adultId) {
		return ErrUserNotFound
	}

	if err := db.Where("user_id = ?", adultId).Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAdultResponsible(tx *gorm.DB, adultId string) (AdultResponsible, error) {
	db := s.dbOrTx(tx)

	adult := AdultResponsible{}
	res := db.Where("responsible_id = ?", adultId).First(&adult)
	if res.RecordNotFound() {
		return AdultResponsible{}, ErrUserNotFound
	}
	if err := res.Error; err != nil {
		return AdultResponsible{}, err
	}

	return adult, nil
}

func (s *Store) UpdateAdultResponsible(tx *gorm.DB, adult AdultResponsible) (AdultResponsible, error) {
	db := s.dbOrTx(tx)

	userId := adult.ResponsibleId
	email := adult.Email

	user := User{}

	if email != "" {
		res := db.Model(&user).Where("user_id = ?", userId).Update("email", email)
		if res.RecordNotFound() {
			return AdultResponsible{}, ErrUserNotFound
		}
		if err := res.Error; err != nil {
			return AdultResponsible{}, err
		}
	}

	adult.ResponsibleId = ""
	adult.Email = ""

	res := db.Where("responsible_id = ?", userId).Model(&AdultResponsible{}).Updates(adult).First(&adult)
	if err := res.Error; err != nil {
		return AdultResponsible{}, err
	}
	if res.RecordNotFound() {
		return AdultResponsible{}, ErrUserNotFound
	}

	return adult, nil
}
