# Объединение модулей проекта «Сенсорный навигатор»

Этот документ описывает, как индивидуальные модули двух разработчиков
команды объединяются в единый проект, расположенный в папке `Project/`.

## Вход и выход интеграции

| Что | Откуда | Куда |
|-----|--------|------|
| Модуль карты и мест | `Project_Атаханов/` | `Project/backend/internal/{handlers,services,models}/places*` + `Project/frontend/src/views/MapView.vue`, `AddPlaceView.vue`, `PlaceView.vue` |
| Модуль пользователей и отзывов | `Project_Насрулаев/` | `Project/backend/internal/{handlers,services,models,auth,middleware}/{auth,users,reviews}*` + `Project/frontend/src/views/{Login,Register,Profile,Favorites}View.vue` |

## Зоны ответственности

```
                     ┌────────────────────────┐
                     │  Vue 3 SPA (frontend)  │
                     │                        │
                     │  ┌──────────────────┐  │
                     │  │   MapView.vue    │  │ ← Атаханов
                     │  │   AddPlaceView   │  │ ← Атаханов
                     │  │   PlaceView      │  │ ← Атаханов
                     │  ├──────────────────┤  │
                     │  │   LoginView      │  │ ← Насрулаев
                     │  │   RegisterView   │  │ ← Насрулаев
                     │  │   ProfileView    │  │ ← Насрулаев
                     │  │   FavoritesView  │  │ ← Насрулаев
                     │  └──────────────────┘  │
                     │   stores/auth.ts       │ ← Насрулаев
                     │   stores/places.ts     │ ← Атаханов
                     │   services/api.ts      │ ← Атаханов
                     └────────────┬───────────┘
                                  │ HTTPS / JSON
                     ┌────────────┴───────────┐
                     │   Go backend (Gin)     │
                     │                        │
                     │  routes/routes.go      │ ← Атаханов
                     │  middleware/auth.go    │ ← Насрулаев
                     │  ┌───────────┐ ┌─────┐ │
                     │  │ places_*  │ │ rev │ │
                     │  │ services  │ │ ews │ │
                     │  │ handlers  │ │  +  │ │
                     │  │           │ │ aut │ │
                     │  │           │ │ h+  │ │
                     │  │           │ │ use │ │
                     │  │           │ │ rs  │ │
                     │  └───────────┘ └─────┘ │
                     │   ↑              ↑     │
                     │   Атаханов  Насрулаев  │
                     └────────────┬───────────┘
                                  │
                          PostgreSQL 16 + PostGIS
                            users / reviews  ← Насрулаев
                            favorites        ← Насрулаев
                            places / categories ← Атаханов
```

## Точки интеграции (контракты между модулями)

Между модулями зафиксировано три контракта.

### 1. Middleware авторизации

**Поставщик**: модуль пользователей и отзывов (Насрулаев Ш. М.)
**Потребитель**: модуль карты и мест (Атаханов Н. Р.)

В индивидуальной сборке Атаханова эндпоинты `POST/PUT/DELETE /api/places`
открыты для всех. При объединении они защищаются middleware:

```go
// internal/routes/routes.go (объединённая версия):
auth := api.Group("/")
auth.Use(middleware.RequireAuth(cfg))
{
    auth.POST("/places", placesH.Create)
    auth.PUT("/places/:id", placesH.Update)
    auth.DELETE("/places/:id", placesH.Delete)
}
```

Идентификатор автора места извлекается из контекста gin (заполняется
`RequireAuth`) и сохраняется в поле `Place.CreatedByID`.

### 2. Агрегированные сенсорные оценки места

**Поставщик**: модуль пользователей и отзывов (Насрулаев Ш. М.)
**Потребитель**: модуль карты и мест (Атаханов Н. Р.)

В индивидуальной сборке модуля карты карточка места показывает только
название, адрес и категорию. После интеграции `places_handler.go` и
`places_service.go` подгружают агрегированные оценки прямо в SQL-запросе
к таблице `reviews` (она существует в общей БД):

```go
selectStmt := strings.Join([]string{
    "places.*",
    "COALESCE(AVG(reviews.noise),0)  AS avg_noise",
    "COALESCE(AVG(reviews.light),0)  AS avg_light",
    "COALESCE(AVG(reviews.crowd),0)  AS avg_crowd",
    "COALESCE(AVG(reviews.smell),0)  AS avg_smell",
    "COALESCE(AVG(reviews.visual),0) AS avg_visual",
    "COUNT(reviews.id) AS reviews_cnt",
}, ", ")
```

Альтернатива (в случае физического разделения сервисов):
HTTP-запрос `GET /api/places/{id}/aggregate`, который реализован в
`Project_Насрулаев/`.

### 3. Общая таблица `places`

В индивидуальной сборке Насрулаева есть упрощённая stub-таблица `places`
(только `id` + `name`), создаваемая `database.SeedDemoPlaces`. При
объединении эта stub-таблица **заменяется** полноценной таблицей мест
из модуля карты со всеми колонками (координаты, адрес, описание,
PostGIS-локация). Внешние ключи `reviews.place_id` и `favorites.place_id`
автоматически работают с новой таблицей, так как имя и тип поля `id`
совпадают.

## Как объединить два индивидуальных проекта

### Вариант 1: уже объединено (рекомендуемый)

Папка `Project/` уже содержит результат интеграции. Это и есть продукт,
который защищается на демо.

### Вариант 2: ручное объединение с нуля

Если нужно повторить процесс merge с двух чистых индивидуальных проектов:

```bash
# 1. Создаём пустой объединённый репозиторий
mkdir Project_Merged
cd Project_Merged
git init

# 2. Заливаем код модуля карты как стартовую базу
cp -r ../Project_Атаханов/* .
git add . && git commit --trailer "Made-with: Cursor" -m "feat: модуль карты и мест (Атаханов Н. Р.)"

# 3. Добавляем модуль пользователей и отзывов
cp -r ../Project_Насрулаев/backend/internal/auth         backend/internal/auth
cp -r ../Project_Насрулаев/backend/internal/middleware   backend/internal/middleware
cp ../Project_Насрулаев/backend/internal/services/users.go    backend/internal/services/
cp ../Project_Насрулаев/backend/internal/services/reviews.go  backend/internal/services/
cp ../Project_Насрулаев/backend/internal/handlers/auth_handler.go    backend/internal/handlers/
cp ../Project_Насрулаев/backend/internal/handlers/users_handler.go   backend/internal/handlers/
cp ../Project_Насрулаев/backend/internal/handlers/reviews_handler.go backend/internal/handlers/
cp -r ../Project_Насрулаев/frontend/src/views/{Login,Register,Profile,Favorites}View.vue \
      frontend/src/views/
cp ../Project_Насрулаев/frontend/src/stores/auth.ts frontend/src/stores/
git add . && git commit --trailer "Made-with: Cursor" -m "feat: модуль пользователей и отзывов (Насрулаев Ш. М.)"

# 4. Объединяем модели и миграции
# В backend/internal/models/models.go должны быть все 5 типов:
#   User, Category, Place, Review, Favorite
# В backend/internal/database/database.go AutoMigrate должен включать все 5 моделей.
git add . && git commit --trailer "Made-with: Cursor" -m "merge: общая схема БД (5 таблиц)"

# 5. Объединяем маршруты
# В backend/internal/routes/routes.go объединяются все эндпоинты:
#   - /api/auth/* (открытые)
#   - /api/places, /api/categories (открытые на чтение)
#   - защищённая группа с RequireAuth для записи и пользовательских эндпоинтов
git add . && git commit --trailer "Made-with: Cursor" -m "merge: единый router с защищёнными группами"

# 6. Сливаем frontend
# В App.vue и router/index.ts должны быть все 7 экранов.
# stores/places.ts (Атаханов) и stores/auth.ts (Насрулаев) живут параллельно.
# services/api.ts объединяется: prefix /api общий, токен подкладывается из stores/auth.
git add . && git commit --trailer "Made-with: Cursor" -m "merge: единый SPA (7 экранов)"

# 7. Расширяем агрегаты
# В services/places.go добавляется JOIN на reviews и AVG-агрегаты,
# Place в models.go получает поля avg_noise, avg_light, avg_crowd,
# avg_smell, avg_visual, overall_avg, reviews_count.
git add . && git commit --trailer "Made-with: Cursor" -m "merge: интеграция агрегатов отзывов в карточку места"
```

## Как проверить корректность объединения

```bash
cd Project/backend
go mod tidy
go build ./...
go vet ./...
go test ./...
```

В корректно объединённом проекте проходят:
- `internal/auth` (3 теста: TestPasswordHashing, TestJWTRoundtrip, TestJWTRejectsWrongSecret)
- `internal/services` (4 теста: TestHaversineKnownDistances + 3 подтеста, TestPlaceFilterDefaults)
- `tests` (2 интеграционных: TestCategoriesEndpoint, TestPlaceCRUD)

Frontend:

```bash
cd Project/frontend
npm install
npm run build
```

## Кто что делал — для отчётности

| Файл / каталог | Автор |
|----------------|-------|
| `backend/internal/models/Place`, `Category` | Атаханов Н. Р. |
| `backend/internal/models/User`, `Review`, `Favorite` | Насрулаев Ш. М. |
| `backend/internal/services/places.go` | Атаханов Н. Р. |
| `backend/internal/services/users.go` | Насрулаев Ш. М. |
| `backend/internal/services/reviews.go` | Насрулаев Ш. М. |
| `backend/internal/handlers/places_handler.go` | Атаханов Н. Р. |
| `backend/internal/handlers/auth_handler.go` | Насрулаев Ш. М. |
| `backend/internal/handlers/users_handler.go` | Насрулаев Ш. М. |
| `backend/internal/handlers/reviews_handler.go` | Насрулаев Ш. М. |
| `backend/internal/auth/jwt.go` | Насрулаев Ш. М. |
| `backend/internal/middleware/auth.go` | Насрулаев Ш. М. |
| `backend/internal/database/database.go` (PostGIS, places) | Атаханов Н. Р. |
| `backend/internal/database/database.go` (users, reviews seed) | Насрулаев Ш. М. |
| `backend/internal/routes/routes.go` | совместная сборка |
| `frontend/src/views/MapView.vue`, `AddPlaceView.vue`, `PlaceView.vue` | Атаханов Н. Р. |
| `frontend/src/views/LoginView.vue`, `RegisterView.vue`, `ProfileView.vue`, `FavoritesView.vue` | Насрулаев Ш. М. |
| `frontend/src/components/FilterPanel.vue`, `PlaceCard.vue`, `SensoryRating.vue` | Атаханов Н. Р. |
| `frontend/src/stores/places.ts` | Атаханов Н. Р. |
| `frontend/src/stores/auth.ts` | Насрулаев Ш. М. |
| `frontend/src/services/api.ts` | Атаханов Н. Р. |
| `frontend/src/router/index.ts`, `App.vue` | Атаханов Н. Р. |
| `backend/docs/openapi.yaml` | Атаханов Н. Р. |
| `docker-compose.yml` | Атаханов Н. Р. |
