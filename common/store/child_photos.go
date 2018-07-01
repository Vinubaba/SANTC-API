package store

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

type ChildPhoto struct {
	PhotoId         sql.NullString
	ChildId         sql.NullString
	PublishedBy     sql.NullString
	ApprovedBy      sql.NullString
	ImageUri        sql.NullString
	Approved        bool
	PublicationDate time.Time
}

func (s *Store) AddChildPhoto(tx *gorm.DB, childPhoto ChildPhoto) error {
	db := s.dbOrTx(tx)

	childPhoto.PhotoId = s.newId()
	if err := db.Create(&childPhoto).Error; err != nil {
		return err
	}
	return nil
}

func (s *Store) ApprovePhoto(tx *gorm.DB, photoId, approvedBy string) error {
	db := s.dbOrTx(tx)
	return db.Table("child_photos").
		Where("photo_id = ?", photoId).
		Update(map[string]interface{}{"approved": true, "approvedBy": approvedBy}).Error
}

func (s *Store) ListPhotos(tx *gorm.DB, options ChildPhotosSearchOptions) ([]ChildPhoto, error) {
	ret := make([]ChildPhoto, 0)

	db := s.dbOrTx(tx)
	query := db.Table("child_photos, children").Select(
		"child_photos.photo_id," +
		"child_photos.child_id," +
		"child_photos.published_by," +
		"child_photos.approved_by," +
		"child_photos.image_uri," +
		"child_photos.approved")
	query = query.Where("child_photos.child_id = children.child_id")
	if options.Approved {
		query = query.Where("child_photos.approved = true")
	} else {
		query = query.Where("child_photos.approved = false")
	}
	if options.DaycareId != "" {
		query = query.Where("children.daycare_id = ?", options.DaycareId)
	}

	rows, err := query.Rows()
	if err != nil {
		return ret, err
	}
	ret, err = s.scanChildPhotosRows(rows)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

type ChildPhotosSearchOptions struct {
	Approved bool
	DaycareId string
}

func (s *Store) scanChildPhotosRows(rows *sql.Rows) ([]ChildPhoto, error) {
	photos := make([]ChildPhoto, 0)
	for rows.Next() {
		currentPhoto := ChildPhoto{}
		if err := rows.Scan(&currentPhoto.PhotoId,
			&currentPhoto.ChildId,
			&currentPhoto.PublishedBy,
			&currentPhoto.ApprovedBy,
			&currentPhoto.ImageUri,
			&currentPhoto.Approved,
		); err != nil {
			return []ChildPhoto{}, err
		}
	}
	return photos, nil
}