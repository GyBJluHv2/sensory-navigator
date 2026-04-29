package services

import (
	"errors"
	"strings"
	"time"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/models"
	"gorm.io/gorm"
)

type PlaceService struct {
	db *gorm.DB
}

func NewPlaceService(db *gorm.DB) *PlaceService {
	return &PlaceService{db: db}
}

// PlaceFilter ????????? ???????? ??? ??????? ????.
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

// List ?????????? ????? ? ??????????????? ????????.
func (s *PlaceService) List(filter PlaceFilter) ([]models.Place, error) {
	if filter.Limit == 0 || filter.Limit > 500 {
		filter.Limit = 200
	}

	// ?????????? ????????? ??? ???????????? JOIN-??????????.
	// ???? Place ?? ????? ????????? ??????? places + ????????, ???????
	// ? ????? Place ???????? gorm:"-" ? ?????? ?? ???????? ? Scan ????????.
	type placeRow struct {
		ID          uint64    `gorm:"column:id"`
		Name        string    `gorm:"column:name"`
		Address     string    `gorm:"column:address"`
		Description string    `gorm:"column:description"`
		CategoryID  uint64    `gorm:"column:category_id"`
		Latitude    float64   `gorm:"column:latitude"`
		Longitude   float64   `gorm:"column:longitude"`
		CreatedByID uint64    `gorm:"column:created_by_id"`
		CreatedAt   time.Time `gorm:"column:created_at"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
		AvgNoise    float64   `gorm:"column:avg_noise"`
		AvgLight    float64   `gorm:"column:avg_light"`
		AvgCrowd    float64   `gorm:"column:avg_crowd"`
		AvgSmell    float64   `gorm:"column:avg_smell"`
		AvgVisual   float64   `gorm:"column:avg_visual"`
		OverallAvg  float64   `gorm:"column:overall_avg"`
		ReviewsCnt  int       `gorm:"column:reviews_cnt"`
	}

	q := s.db.Table("places").
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

	var rows []placeRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	places := make([]models.Place, 0, len(rows))
	for _, r := range rows {
		p := models.Place{
			ID:          r.ID,
			Name:        r.Name,
			Address:     r.Address,
			Description: r.Description,
			CategoryID:  r.CategoryID,
			Latitude:    r.Latitude,
			Longitude:   r.Longitude,
			CreatedByID: r.CreatedByID,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			AvgNoise:    r.AvgNoise,
			AvgLight:    r.AvgLight,
			AvgCrowd:    r.AvgCrowd,
			AvgSmell:    r.AvgSmell,
			AvgVisual:   r.AvgVisual,
			OverallAvg:  r.OverallAvg,
			ReviewsCnt:  r.ReviewsCnt,
		}
		if p.CategoryID != 0 {
			var cat models.Category
			if err := s.db.First(&cat, p.CategoryID).Error; err == nil {
				p.Category = cat
			}
		}
		places = append(places, p)
	}

	return places, nil
}

// Get ?????????? ????? ?? ??????????????, ??????? ?????????????? ??????.
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

// Nearby ?????????? ????? ? ???????? ??????? (? ??????).
// ???? PostGIS ??????????, ???????????? ST_DWithin (????? ??????? location).
// ????? ??????????? ??????????? ??????? ????????? — ?????????? ??????.
func (s *PlaceService) Nearby(lat, lon float64, radiusMeters int) ([]models.Place, error) {
	if radiusMeters <= 0 {
		return nil, errors.New("?????? ?????? ???? ?????????????")
	}

	var places []models.Place

	// ??????? ??????? PostGIS, ???? ?????????? ??????????? ? ??????? location ??????????.
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

	// Fallback: ?????? ????? ??????? ????????? (R = 6371000 ?).
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

// Create ??????? ????? ?????.
func (s *PlaceService) Create(p *models.Place) error {
	return s.db.Create(p).Error
}

// Update ????????? ?????????? ?????.
func (s *PlaceService) Update(p *models.Place) error {
	return s.db.Save(p).Error
}

// Delete ??????? ????? ?? id.
func (s *PlaceService) Delete(id uint64) error {
	return s.db.Delete(&models.Place{}, id).Error
}

// Categories ?????????? ??? ????????? ?????????.
func (s *PlaceService) Categories() ([]models.Category, error) {
	var cats []models.Category
	err := s.db.Order("name").Find(&cats).Error
	return cats, err
}