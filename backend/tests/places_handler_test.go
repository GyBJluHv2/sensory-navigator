package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/config"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/database"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/models"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

// Тесты HTTP-маршрутов модуля карт и мест.
// Активируются при наличии переменных окружения с настройками тестовой БД
// PostgreSQL/PostGIS. По умолчанию пропускаются (для CI без БД).

func setupRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	if os.Getenv("TEST_DB_DSN") == "" {
		t.Skip("TEST_DB_DSN не задан — тесты HTTP пропущены")
	}
	gin.SetMode(gin.TestMode)
	cfg := config.Load()
	cfg.JWTSecret = "test-secret"
	db, err := gorm.Open(postgres.Open(os.Getenv("TEST_DB_DSN")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))
	require.NoError(t, database.SeedCategories(db))
	r := routes.NewRouter(db, cfg)
	return r, db
}

func TestCategoriesEndpoint(t *testing.T) {
	r, _ := setupRouter(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Items []models.Category `json:"items"`
		Count int               `json:"count"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.GreaterOrEqual(t, body.Count, 5)
}

func TestPlaceCRUD(t *testing.T) {
	r, db := setupRouter(t)

	// Регистрируем пользователя и получаем токен
	regBody, _ := json.Marshal(map[string]any{
		"email": "tester@example.com", "username": "tester",
		"password": "tester123", "display_name": "Тест",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusConflict)

	loginBody, _ := json.Marshal(map[string]string{
		"email": "tester@example.com", "password": "tester123",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var login struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &login))
	require.NotEmpty(t, login.Token)

	// Создаём место
	var cats []models.Category
	db.Find(&cats)
	require.NotEmpty(t, cats)

	createBody, _ := json.Marshal(map[string]any{
		"name": "Тестовое кафе", "address": "ул. Тестовая, 1",
		"description": "test", "category_id": cats[0].ID,
		"latitude": 55.7558, "longitude": 37.6173,
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/places",
		bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+login.Token)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created models.Place
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.NotZero(t, created.ID)

	// Получаем место по id
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet,
		"/api/places/"+itoa(created.ID), nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Поиск по имени
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/places/search?q=Тестовое", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Поиск рядом
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet,
		"/api/places/nearby?lat=55.7558&lon=37.6173&radius=2000", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Удаляем место
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete,
		"/api/places/"+itoa(created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+login.Token)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func itoa(n uint64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = digits[n%10]
		n /= 10
	}
	return string(buf[pos:])
}
