package store

import "context"

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

func (s *Store) AddAdultResponsible(ctx context.Context, adult AdultResponsible) (AdultResponsible, error) {
	if err := s.Db.Create(&adult).Error; err != nil {
		return AdultResponsible{}, err
	}

	return adult, nil
}

func (s *Store) ListAdultResponsible(ctx context.Context) ([]AdultResponsible, error) {
	adults := []AdultResponsible{}
	if err := s.Db.Find(&adults).Error; err != nil {
		return nil, err
	}

	return adults, nil
}

func (s *Store) DeleteAdultResponsible(ctx context.Context, adultId string) (err error) {
	if !s.userExists(ctx, adultId) {
		return ErrUserNotFound
	}

	if err := s.Db.Where("user_id = ?", adultId).Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAdultResponsible(ctx context.Context, adultId string) (AdultResponsible, error) {
	adult := AdultResponsible{}
	res := s.Db.Where("responsible_id = ?", adultId).First(&adult)
	if res.RecordNotFound() {
		return AdultResponsible{}, ErrUserNotFound
	}
	if err := res.Error; err != nil {
		return AdultResponsible{}, err
	}

	return adult, nil
}

func (s *Store) UpdateAdultResponsible(ctx context.Context, adult AdultResponsible) (AdultResponsible, error) {
	userId := adult.ResponsibleId
	email := adult.Email

	user := User{}

	if email != "" {
		res := s.Db.Model(&user).Where("user_id = ?", userId).Update("email", email)
		if res.RecordNotFound() {
			return AdultResponsible{}, ErrUserNotFound
		}
		if err := res.Error; err != nil {
			return AdultResponsible{}, err
		}
	}

	adult.ResponsibleId = ""
	adult.Email = ""

	res := s.Db.Where("responsible_id = ?", userId).Model(&AdultResponsible{}).Updates(adult).First(&adult)
	if err := res.Error; err != nil {
		return AdultResponsible{}, err
	}
	if res.RecordNotFound() {
		return AdultResponsible{}, ErrUserNotFound
	}

	return adult, nil
}
