package store

import (
	"context"
	"errors"
	"time"
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

func (s *Store) AddChild(ctx context.Context, child Child) (Child, error) {
	child.ChildId = s.StringGenerator.GenerateUuid()

	if err := s.Db.Create(&child).Error; err != nil {
		return Child{}, err
	}

	return child, nil
}

func (s *Store) DeleteChild(ctx context.Context, childId string) (err error) {
	if !s.childExists(ctx, childId) {
		return ErrChildNotFound
	}

	if err := s.Db.Where("child_id = ?", childId).Delete(&Child{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) childExists(ctx context.Context, childId string) bool {
	c := Child{ChildId: childId}
	return !s.Db.Model(Child{}).Where("child_id = ?", childId).First(&c).RecordNotFound()
}

func (s *Store) GetChild(ctx context.Context, childId string) (Child, error) {
	child := Child{}
	res := s.Db.Where("child_id = ?", childId).First(&child)
	if res.RecordNotFound() {
		return Child{}, ErrChildNotFound
	}
	if err := res.Error; err != nil {
		return Child{}, err
	}

	return child, nil
}

func (s *Store) ListChild(ctx context.Context) ([]Child, error) {
	children := []Child{}
	if err := s.Db.Find(&children).Error; err != nil {
		return nil, err
	}

	return children, nil
}

func (s *Store) UpdateChild(ctx context.Context, child Child) (Child, error) {
	res := s.Db.Where("child_id = ?", child.ChildId).Model(&Child{}).Updates(child).First(&child)
	if res.RecordNotFound() {
		return Child{}, ErrChildNotFound
	}
	if err := res.Error; err != nil {
		return Child{}, err
	}

	return child, nil
}
