package services

import (
	"errors"
	"strings"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/models"
	"gorm.io/gorm"
)

type PlaceService struct {
	db *gorm.DB
}

func NewPlaceService(db *gorm.DB) *PlaceService {
	return &PlaceService{db: db}
}

// PlaceFilter описывает критерии для выборки мест.
type PlaceFilter struct {
	CategoryID uint64
	Search     string
	NoiseMax   int
	LightMax   int
	CrowdMax   int
	SmellMax   int
	VisualMax  int
	Limit      int
	Offset     int
}

// List возвращает места с агрегированными оценками.
func (s *PlaceService) List(filter PlaceFilter) ([]models.Place, error) {
	if filter.Limit == 0 || filter.Limit > 500 {
		filter.Limit = 200
	}

	q := s.db.Model(&models.Place{}).
		Preload("Category").
		Joins("LEFT JOIN reviews ON reviews.place_id = places.id").
		Group("places.id")

	selectStmt := strings.Join([]string{
		"places.*",
		"COALESCE(AVG(reviews.noise),0) AS avg_noise",
		"COALESCE(AVG(reviews.light),0) AS avg_light",
		"COALESCE(AVG(reviews.crowd),0) AS avg_crowd",
		"COALESCE(AVG(reviews.smell),0) AS avg_smell",
		"COALESCE(AVG(reviews.visual),0) AS avg_visual",
		"COALESCE(AVG((reviews.noise+reviews.light+reviews.crowd+reviews.smell+reviews.visual)/5.0),0) AS overall_avg",
		"COUNT(reviews.id) AS reviews_cnt",
	}, ", ")
	q = q.Select(selectStmt)

	if filter.CategoryID != 0 {
		q = q.Where("places.category_id = ?", filter.CategoryID)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(places.name) LIKE ? OR LOWER(places.address) LIKE ?", like, like)
	}

	if filter.NoiseMax > 0 {
		q = q.Having("COALESCE(AVG(reviews.noise),0) <= ?", filter.NoiseMax)
	}
	if filter.LightMax > 0 {
		q = q.Having("COALESCE(AVG(reviews.light),0) <= ?", filter.LightMax)
	}
	if filter.CrowdMax > 0 {
		q = q.Having("COALESCE(AVG(reviews.crowd),0) <= ?", filter.CrowdMax)
	}
	if filter.SmellMax > 0 {
		q = q.Having("COALESCE(AVG(reviews.smell),0) <= ?", filter.SmellMax)
	}
	if filter.VisualMax > 0 {
		q = q.Having("COALESCE(AVG(reviews.visual),0) <= ?", filter.VisualMax)
	}

	q = q.Limit(filter.Limit).Offset(filter.Offset)

	var places []models.Place
	if err := q.Scan(&places).Error; err != nil {
		return nil, err
	}

	for i := range places {
		var cat models.Category
		if err := s.db.First(&cat, places[i].CategoryID).Error; err == nil {
			places[i].Category = cat
		}
	}

	return places, nil
}

// Get возвращает место по идентификатору, включая агрегированные оценки.
func (s *PlaceService) Get(id uint64) (*models.Place, error) {
	var place models.Place
	if err := s.db.Preload("Category").First(&place, id).Error; err != nil {
		return nil, err
	}
	type agg struct {
		AvgNoise   float64
		AvgLight   float64
		AvgCrowd   float64
		AvgSmell   float64
		AvgVisual  float64
		OverallAvg float64
		ReviewsCnt int
	}
	var a agg
	s.db.Model(&models.Review{}).
		Select(`COALESCE(AVG(noise),0) AS avg_noise,
			COALESCE(AVG(light),0) AS avg_light,
			COALESCE(AVG(crowd),0) AS avg_crowd,
			COALESCE(AVG(smell),0) AS avg_smell,
			COALESCE(AVG(visual),0) AS avg_visual,
			COALESCE(AVG((noise+light+crowd+smell+visual)/5.0),0) AS overall_avg,
			COUNT(*) AS reviews_cnt`).
		Where("place_id = ?", id).
		Scan(&a)
	place.AvgNoise = a.AvgNoise
	place.AvgLight = a.AvgLight
	place.AvgCrowd = a.AvgCrowd
	place.AvgSmell = a.AvgSmell
	place.AvgVisual = a.AvgVisual
	place.OverallAvg = a.OverallAvg
	place.ReviewsCnt = a.ReviewsCnt
	return &place, nil
}

// Nearby возвращает места в заданном радиусе (в метрах).
// Если PostGIS установлен, используется ST_DWithin (через колонку location).
// Иначе применяется сферическая формула Хаверсина — встроенный фолбэк.
func (s *PlaceService) Nearby(lat, lon float64, radiusMeters int) ([]models.Place, error) {
	if radiusMeters <= 0 {
		return nil, errors.New("радиус должен быть положительным")
	}

	var places []models.Place

	// Сначала пробуем PostGIS, если расширение установлено и колонка location существует.
	postGISErr := s.db.Raw(
		`SELECT places.*, c.id AS "category__id", c.name AS "category__name",
		        c.slug AS "category__slug", c.icon AS "category__icon"
		   FROM places
		   LEFT JOIN categories c ON c.id = places.category_id
		  WHERE ST_DWithin(places.location,
		                   ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
		                   ?)
		  ORDER BY ST_Distance(places.location,
		                       ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography)
		  LIMIT 200`,
		lon, lat, radiusMeters, lon, lat,
	).Scan(&places).Error
	if postGISErr == nil && len(places) >= 0 && tablesUseLocation(s.db) {
		s.fillCategories(places)
		s.fillAggregates(places)
		return places, nil
	}

	// Fallback: расчёт через формулу Хаверсина (R = 6371000 м).
	q := s.db.Raw(
		`SELECT * FROM places WHERE
		   6371000 * acos(
		     LEAST(1.0,
		       cos(radians(?)) * cos(radians(latitude)) *
		       cos(radians(longitude) - radians(?)) +
		       sin(radians(?)) * sin(radians(latitude))
		     )
		   ) <= ?
		 ORDER BY 6371000 * acos(
		     LEAST(1.0,
		       cos(radians(?)) * cos(radians(latitude)) *
		       cos(radians(longitude) - radians(?)) +
		       sin(radians(?)) * sin(radians(latitude))
		     )
		   ) ASC
		 LIMIT 200`,
		lat, lon, lat, radiusMeters, lat, lon, lat,
	)
	if err := q.Scan(&places).Error; err != nil {
		return nil, err
	}
	s.fillCategories(places)
	s.fillAggregates(places)
	return places, nil
}

func tablesUseLocation(db *gorm.DB) bool {
	var n int
	db.Raw(`SELECT 1 FROM information_schema.columns
		WHERE table_name='places' AND column_name='location' LIMIT 1`).Scan(&n)
	return n == 1
}

func (s *PlaceService) fillCategories(places []models.Place) {
	for i := range places {
		if places[i].CategoryID == 0 {
			continue
		}
		var c models.Category
		if err := s.db.First(&c, places[i].CategoryID).Error; err == nil {
			places[i].Category = c
		}
	}
}

func (s *PlaceService) fillAggregates(places []models.Place) {
	for i := range places {
		var a struct {
			AvgNoise, AvgLight, AvgCrowd, AvgSmell, AvgVisual, OverallAvg float64
			ReviewsCnt                                                    int
		}
		s.db.Model(&models.Review{}).
			Select(`COALESCE(AVG(noise),0) AS avg_noise,
				COALESCE(AVG(light),0) AS avg_light,
				COALESCE(AVG(crowd),0) AS avg_crowd,
				COALESCE(AVG(smell),0) AS avg_smell,
				COALESCE(AVG(visual),0) AS avg_visual,
				COALESCE(AVG((noise+light+crowd+smell+visual)/5.0),0) AS overall_avg,
				COUNT(*) AS reviews_cnt`).
			Where("place_id = ?", places[i].ID).
			Scan(&a)
		places[i].AvgNoise = a.AvgNoise
		places[i].AvgLight = a.AvgLight
		places[i].AvgCrowd = a.AvgCrowd
		places[i].AvgSmell = a.AvgSmell
		places[i].AvgVisual = a.AvgVisual
		places[i].OverallAvg = a.OverallAvg
		places[i].ReviewsCnt = a.ReviewsCnt
	}
}

// Create создаёт новое место.
func (s *PlaceService) Create(p *models.Place) error {
	return s.db.Create(p).Error
}

// Update сохраняет изменённое место.
func (s *PlaceService) Update(p *models.Place) error {
	return s.db.Save(p).Error
}

// Delete удаляет место по id.
func (s *PlaceService) Delete(id uint64) error {
	return s.db.Delete(&models.Place{}, id).Error
}

// Categories возвращает все доступные категории.
func (s *PlaceService) Categories() ([]models.Category, error) {
	var cats []models.Category
	err := s.db.Order("name").Find(&cats).Error
	return cats, err
}