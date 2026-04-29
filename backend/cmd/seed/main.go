package main

// Seed загружает базу демо-данными: пользователь demo@example.com и набор мест
// в центре Москвы по разным категориям. Запуск: go run ./cmd/seed.

import (
	"log"

	"github.com/atakhanov/sensory-navigator/backend/internal/auth"
	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/atakhanov/sensory-navigator/backend/internal/database"
	"github.com/atakhanov/sensory-navigator/backend/internal/models"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}
	if err := database.SeedCategories(db); err != nil {
		log.Fatalf("ошибка категорий: %v", err)
	}

	cats := map[string]uint64{}
	var all []models.Category
	db.Find(&all)
	for _, c := range all {
		cats[c.Slug] = c.ID
	}

	hash, _ := auth.GeneratePasswordHash("demo123", cfg.BCryptCost)
	user := &models.User{
		Email:        "demo@example.com",
		Username:     "demo",
		PasswordHash: hash,
		DisplayName:  "Демо-пользователь",
		NoisePref:    2, LightPref: 2, CrowdPref: 2,
	}
	db.Where("email = ?", user.Email).FirstOrCreate(user)

	demoPlaces := []models.Place{
		{Name: "Городская библиотека на Малой Дмитровке", Address: "Москва, Малая Дмитровка, 8/1",
			CategoryID: cats["library"], Latitude: 55.7702, Longitude: 37.6028,
			Description: "Спокойная читальня с приглушённым светом и зонами тишины."},
		{Name: "Парк Горького", Address: "Москва, ул. Крымский Вал, 9",
			CategoryID: cats["park"], Latitude: 55.7298, Longitude: 37.6019,
			Description: "Большой парк с зелёными зонами и набережной."},
		{Name: "Кофейня «Тихий уголок»", Address: "Москва, ул. Покровка, 15",
			CategoryID: cats["cafe"], Latitude: 55.7588, Longitude: 37.6447,
			Description: "Маленькая кофейня без музыки, с уютным светом."},
		{Name: "ТЦ «Авиапарк»", Address: "Москва, Ходынский бульвар, 4",
			CategoryID: cats["mall"], Latitude: 55.7889, Longitude: 37.5306,
			Description: "Один из крупнейших ТЦ — высокая нагрузка по шуму и людности."},
		{Name: "Третьяковская галерея", Address: "Москва, Лаврушинский пер., 10",
			CategoryID: cats["museum"], Latitude: 55.7415, Longitude: 37.6209,
			Description: "Главный филиал галереи, спокойная атмосфера."},
		{Name: "Парк «Зарядье»", Address: "Москва, ул. Варварка, 6",
			CategoryID: cats["park"], Latitude: 55.7515, Longitude: 37.6294,
			Description: "Современный парк рядом с Красной площадью."},
		{Name: "Кинотеатр «Каро»", Address: "Москва, ул. Земляной Вал, 33",
			CategoryID: cats["cinema"], Latitude: 55.7569, Longitude: 37.6573,
			Description: "Стандартный кинотеатр с обычным уровнем шума и света."},
	}
	for _, p := range demoPlaces {
		var existing models.Place
		err := db.Where("name = ?", p.Name).First(&existing).Error
		if err != nil {
			p.CreatedByID = user.ID
			db.Create(&p)
		}
	}

	log.Printf("seed выполнен: пользователь=%s, мест=%d", user.Email, len(demoPlaces))
}
