package database

import (
	"errors"
	"log"

	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/atakhanov/sensory-navigator/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := enablePostGIS(db); err != nil {
		log.Printf("PostGIS недоступен (запрос nearby будет использовать формулу Хаверсина): %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Place{},
		&models.Review{},
		&models.Favorite{},
	); err != nil {
		return err
	}

	createGISTIndex(db)

	return nil
}

func enablePostGIS(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error
}

// createGISTIndex добавляет geography-колонку и пространственный индекс,
// если PostGIS установлен. Подобные манипуляции — единственный нетривиальный
// и не покрываемый миграциями GORM участок логики на SQL.
func createGISTIndex(db *gorm.DB) {
	statements := []string{
		`ALTER TABLE places ADD COLUMN IF NOT EXISTS location GEOGRAPHY(POINT, 4326)`,
		`CREATE INDEX IF NOT EXISTS idx_places_location ON places USING GIST(location)`,
		`CREATE OR REPLACE FUNCTION sync_place_location() RETURNS trigger AS $$
		 BEGIN
		   NEW.location = ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326)::geography;
		   RETURN NEW;
		 END;
		 $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS trg_sync_place_location ON places`,
		`CREATE TRIGGER trg_sync_place_location
		   BEFORE INSERT OR UPDATE OF latitude, longitude ON places
		   FOR EACH ROW EXECUTE FUNCTION sync_place_location()`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("настройка PostGIS пропущена: %v", err)
			return
		}
	}
}

var defaultCategories = []models.Category{
	{Name: "Кафе и рестораны", Slug: "cafe", Description: "Заведения общественного питания", Icon: "cafe"},
	{Name: "Библиотеки", Slug: "library", Description: "Тихие места для чтения и работы", Icon: "library"},
	{Name: "Парки и скверы", Slug: "park", Description: "Открытые зелёные зоны", Icon: "park"},
	{Name: "Торговые центры", Slug: "mall", Description: "Крупные торговые комплексы", Icon: "mall"},
	{Name: "Музеи и выставки", Slug: "museum", Description: "Культурные пространства", Icon: "museum"},
	{Name: "Кинотеатры", Slug: "cinema", Description: "Помещения для просмотра фильмов", Icon: "cinema"},
	{Name: "Спортивные объекты", Slug: "sport", Description: "Залы, стадионы, бассейны", Icon: "sport"},
	{Name: "Образование", Slug: "edu", Description: "Учебные аудитории и коворкинги", Icon: "edu"},
}

func SeedCategories(db *gorm.DB) error {
	for _, c := range defaultCategories {
		var exists models.Category
		err := db.Where("slug = ?", c.Slug).First(&exists).Error
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			if cerr := db.Create(&c).Error; cerr != nil {
				return cerr
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}