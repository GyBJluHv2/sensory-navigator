package routes

import (
	"net/http"
	"time"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/config"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/handlers"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/middleware"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	placeSvc := services.NewPlaceService(db)
	userSvc := services.NewUserService(db, cfg)
	reviewSvc := services.NewReviewService(db)

	authH := handlers.NewAuthHandler(userSvc, cfg)
	usersH := handlers.NewUsersHandler(userSvc)
	placesH := handlers.NewPlacesHandler(placeSvc)
	reviewsH := handlers.NewReviewsHandler(reviewSvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// Открытые маршруты
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)

	api.GET("/categories", placesH.Categories)

	// Места: чтение — публичное; запись — только для аутентифицированных
	api.GET("/places", placesH.List)
	api.GET("/places/search", placesH.List)
	api.GET("/places/nearby", placesH.Nearby)
	api.GET("/places/:id", placesH.Get)
	api.GET("/places/:id/reviews", reviewsH.ListByPlace)

	// Защищённые маршруты
	auth := api.Group("/")
	auth.Use(middleware.RequireAuth(cfg))
	{
		auth.GET("/me", usersH.Me)
		auth.PUT("/me", usersH.UpdateMe)
		auth.PUT("/me/password", usersH.ChangePassword)

		auth.POST("/places", placesH.Create)
		auth.PUT("/places/:id", placesH.Update)
		auth.DELETE("/places/:id", placesH.Delete)

		auth.POST("/places/:id/reviews", reviewsH.Create)
		auth.PUT("/reviews/:id", reviewsH.Update)
		auth.DELETE("/reviews/:id", reviewsH.Delete)
		auth.GET("/reviews/me", reviewsH.ListMyReviews)

		auth.POST("/places/:id/favorite", reviewsH.AddFavorite)
		auth.DELETE("/places/:id/favorite", reviewsH.RemoveFavorite)
		auth.GET("/favorites", reviewsH.ListFavorites)
	}

	return r
}