package store

import "context"

type OfficeManager struct {
	OfficeManagerId string
	Email         string
}

func (s *Store) AddOfficeManager(ctx context.Context, officeManager OfficeManager) (OfficeManager, error) {
	if err := s.Db.Create(&officeManager).Error; err != nil {
		return OfficeManager{}, err
	}

	return officeManager, nil
}

func (s *Store) ListOfficeManager(ctx context.Context) ([]OfficeManager, error) {
	officeManagers := []OfficeManager{}
	if err := s.Db.Find(&officeManagers).Error; err != nil {
		return nil, err
	}

	return officeManagers, nil
}

func (s *Store) DeleteOfficeManager(ctx context.Context, officeManagerId string) (err error) {
	if !s.userExists(ctx, officeManagerId) {
		return ErrUserNotFound
	}

	if err := s.Db.Where("user_id = ?", officeManagerId).Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOfficeManager(ctx context.Context, officeManagerId string) (OfficeManager, error) {
	officeManager := OfficeManager{}
	res := s.Db.Where("office_manager_id = ?", officeManagerId).First(&officeManager)
	if res.RecordNotFound() {
		return OfficeManager{}, ErrUserNotFound
	}
	if err := res.Error; err != nil {
		return OfficeManager{}, err
	}

	return officeManager, nil
}

func (s *Store) UpdateOfficeManager(ctx context.Context, officeManager OfficeManager) (OfficeManager, error) {
	userId := officeManager.OfficeManagerId
	email := officeManager.Email

	user := User{}

	if email != "" {
		res := s.Db.Model(&user).Where("user_id = ?", userId).Update("email", email)
		if res.RecordNotFound() {
			return OfficeManager{}, ErrUserNotFound
		}
		if err := res.Error; err != nil {
			return OfficeManager{}, err
		}
	}

	officeManager.OfficeManagerId = ""
	officeManager.Email = ""

	res := s.Db.Where("office_manager_id = ?", userId).Model(&OfficeManager{}).Updates(officeManager).First(&officeManager)
	if err := res.Error; err != nil {
		return OfficeManager{}, err
	}
	if res.RecordNotFound() {
		return OfficeManager{}, ErrUserNotFound
	}

	return officeManager, nil
}
