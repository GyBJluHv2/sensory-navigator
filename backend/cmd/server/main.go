package main

import (
	"log"
	"os"

	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/atakhanov/sensory-navigator/backend/internal/database"
	"github.com/atakhanov/sensory-navigator/backend/internal/routes"
	"github.com/joho/godotenv"
)

// @title       Sensory Navigator API
// @version     1.0
// @description REST API сервиса «Сенсорный навигатор» — карты с отзывами
// @description для людей, склонных к сенсорной перегрузке.
// @host        localhost:8080
// @BasePath    /api
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, используются переменные окружения")
	}

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("ошибка миграций схемы БД: %v", err)
	}

	if err := database.SeedCategories(db); err != nil {
		log.Fatalf("ошибка инициализации категорий: %v", err)
	}

	router := routes.NewRouter(db, cfg)

	addr := ":" + cfg.Port
	log.Printf("запуск сервера на %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("ошибка запуска HTTP-сервера: %v", err)
		os.Exit(1)
	}
}