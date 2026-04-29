package services

import (
	"errors"

	"github.com/atakhanov/sensory-navigator/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrReviewExists = errors.New("отзыв этого пользователя для места уже существует")
	ErrNotOwnReview = errors.New("отзыв принадлежит другому пользователю")
)

type ReviewService struct {
	db *gorm.DB
}

func NewReviewService(db *gorm.DB) *ReviewService {
	return &ReviewService{db: db}
}

func (s *ReviewService) Create(r *models.Review) error {
	var n int64
	s.db.Model(&models.Review{}).
		Where("user_id = ? AND place_id = ?", r.UserID, r.PlaceID).
		Count(&n)
	if n > 0 {
		return ErrReviewExists
	}
	return s.db.Create(r).Error
}

func (s *ReviewService) Update(userID uint64, r *models.Review) error {
	var existing models.Review
	if err := s.db.First(&existing, r.ID).Error; err != nil {
		return err
	}
	if existing.UserID != userID {
		return ErrNotOwnReview
	}
	existing.Text = r.Text
	existing.Noise = r.Noise
	existing.Light = r.Light
	existing.Crowd = r.Crowd
	existing.Smell = r.Smell
	existing.Visual = r.Visual
	return s.db.Save(&existing).Error
}

func (s *ReviewService) Delete(userID, reviewID uint64) error {
	var existing models.Review
	if err := s.db.First(&existing, reviewID).Error; err != nil {
		return err
	}
	if existing.UserID != userID {
		return ErrNotOwnReview
	}
	return s.db.Delete(&existing).Error
}

func (s *ReviewService) ListByPlace(placeID uint64) ([]models.Review, error) {
	var rs []models.Review
	err := s.db.Preload("User").
		Where("place_id = ?", placeID).
		Order("created_at DESC").
		Find(&rs).Error
	return rs, err
}

func (s *ReviewService) ListByUser(userID uint64) ([]models.Review, error) {
	var rs []models.Review
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&rs).Error
	return rs, err
}

func (s *ReviewService) AddFavorite(userID, placeID uint64) error {
	fav := models.Favorite{UserID: userID, PlaceID: placeID}
	return s.db.Where("user_id = ? AND place_id = ?", userID, placeID).
		FirstOrCreate(&fav).Error
}

func (s *ReviewService) RemoveFavorite(userID, placeID uint64) error {
	return s.db.Where("user_id = ? AND place_id = ?", userID, placeID).
		Delete(&models.Favorite{}).Error
}

func (s *ReviewService) ListFavorites(userID uint64) ([]models.Place, error) {
	var places []models.Place
	err := s.db.
		Joins("JOIN favorites f ON f.place_id = places.id").
		Where("f.user_id = ?", userID).
		Preload("Category").
		Order("f.created_at DESC").
		Find(&places).Error
	return places, err
}