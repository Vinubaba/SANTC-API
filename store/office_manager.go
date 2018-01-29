package store

import (
	"github.com/jinzhu/gorm"
)

type OfficeManager struct {
	OfficeManagerId string
	Email           string
}

func (s *Store) AddOfficeManager(tx *gorm.DB, officeManager OfficeManager) (OfficeManager, error) {
	db := s.dbOrTx(tx)

	if err := db.Create(&officeManager).Error; err != nil {
		return OfficeManager{}, err
	}

	return officeManager, nil
}

func (s *Store) ListOfficeManager(tx *gorm.DB) ([]OfficeManager, error) {
	db := s.dbOrTx(tx)

	officeManagers := []OfficeManager{}
	if err := db.Find(&officeManagers).Error; err != nil {
		return nil, err
	}

	return officeManagers, nil
}

func (s *Store) DeleteOfficeManager(tx *gorm.DB, officeManagerId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.userExists(db, officeManagerId) {
		return ErrUserNotFound
	}

	if err := db.Where("user_id = ?", officeManagerId).Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOfficeManager(tx *gorm.DB, officeManagerId string) (OfficeManager, error) {
	db := s.dbOrTx(tx)

	officeManager := OfficeManager{}
	res := db.Where("office_manager_id = ?", officeManagerId).First(&officeManager)
	if res.RecordNotFound() {
		return OfficeManager{}, ErrUserNotFound
	}
	if err := res.Error; err != nil {
		return OfficeManager{}, err
	}

	return officeManager, nil
}

func (s *Store) UpdateOfficeManager(tx *gorm.DB, officeManager OfficeManager) (OfficeManager, error) {
	db := s.dbOrTx(tx)

	userId := officeManager.OfficeManagerId
	email := officeManager.Email

	user := User{}

	if email != "" {
		res := db.Model(&user).Where("user_id = ?", userId).Update("email", email)
		if res.RecordNotFound() {
			return OfficeManager{}, ErrUserNotFound
		}
		if err := res.Error; err != nil {
			return OfficeManager{}, err
		}
	}

	officeManager.OfficeManagerId = ""
	officeManager.Email = ""

	res := db.Where("office_manager_id = ?", userId).Model(&OfficeManager{}).Updates(officeManager).First(&officeManager)
	if err := res.Error; err != nil {
		return OfficeManager{}, err
	}
	if res.RecordNotFound() {
		return OfficeManager{}, ErrUserNotFound
	}

	return officeManager, nil
}
