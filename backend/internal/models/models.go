package models

import (
	"time"
)

// User — учётная запись пользователя сервиса.
type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string    `gorm:"size:128" json:"display_name"`
	AvatarURL    string    `gorm:"size:512" json:"avatar_url"`
	NoisePref    int       `gorm:"default:3" json:"noise_pref"`
	LightPref    int       `gorm:"default:3" json:"light_pref"`
	CrowdPref    int       `gorm:"default:3" json:"crowd_pref"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Category — категория места (кафе, библиотека, парк и т. п.).
type Category struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;size:64;not null" json:"slug"`
	Description string    `gorm:"size:500" json:"description"`
	Icon        string    `gorm:"size:64" json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
}

// Place — общественное место с координатами и характеристиками.
// Координаты хранятся как пара широта/долгота (SRID 4326). При наличии PostGIS
// дополнительно поддерживается колонка location типа GEOGRAPHY(POINT, 4326).
type Place struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Address     string    `gorm:"size:500" json:"address"`
	Description string    `gorm:"type:text" json:"description"`
	CategoryID  uint64    `gorm:"index;not null" json:"category_id"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Latitude    float64   `gorm:"not null" json:"latitude"`
	Longitude   float64   `gorm:"not null" json:"longitude"`
	CreatedByID uint64    `gorm:"index" json:"created_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Агрегированные оценки рассчитываются запросом для клиента.
	AvgNoise   float64 `gorm:"-" json:"avg_noise"`
	AvgLight   float64 `gorm:"-" json:"avg_light"`
	AvgCrowd   float64 `gorm:"-" json:"avg_crowd"`
	AvgSmell   float64 `gorm:"-" json:"avg_smell"`
	AvgVisual  float64 `gorm:"-" json:"avg_visual"`
	OverallAvg float64 `gorm:"-" json:"overall_avg"`
	ReviewsCnt int     `gorm:"-" json:"reviews_count"`
}

// Review — отзыв пользователя о месте с сенсорными оценками 1..5.
type Review struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	PlaceID    uint64    `gorm:"index;not null" json:"place_id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Text       string    `gorm:"type:text" json:"text"`
	Noise      int       `gorm:"not null" json:"noise"`
	Light      int       `gorm:"not null" json:"light"`
	Crowd      int       `gorm:"not null" json:"crowd"`
	Smell      int       `gorm:"not null" json:"smell"`
	Visual     int       `gorm:"not null" json:"visual"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Favorite — избранное место пользователя.
type Favorite struct {
	UserID    uint64    `gorm:"primaryKey" json:"user_id"`
	PlaceID   uint64    `gorm:"primaryKey" json:"place_id"`
	CreatedAt time.Time `json:"created_at"`
}