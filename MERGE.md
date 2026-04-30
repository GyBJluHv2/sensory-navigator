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

**Поставщик**: модуль пользователей и отзывов (Насрулаев Шарапудин Махадович)
**Потребитель**: модуль карты и мест (Атаханов Набиюлла Румиевич)

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

**Поставщик**: модуль пользователей и отзывов (Насрулаев Шарапудин Махадович)
**Потребитель**: модуль карты и мест (Атаханов Набиюлла Румиевич)

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
git add . && git commit -m "feat: модуль карты и мест (Атаханов Набиюлла Румиевич)"

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
git add . && git commit -m "feat: модуль пользователей и отзывов (Насрулаев Шарапудин Махадович)"

# 4. Объединяем модели и миграции
# В backend/internal/models/models.go должны быть все основные типы сущностей,
# включая User, Category, Place, Review, Favorite и (при двухшаговой регистрации)
# VerificationCode.
# В backend/internal/database/database.go AutoMigrate должен включать все эти модели.
git add . && git commit -m "merge: общая схема БД и миграции"

# 5. Объединяем маршруты
# В backend/internal/routes/routes.go объединяются все эндпоинты:
#   - /api/auth/* (открытые)
#   - /api/places, /api/categories (открытые на чтение)
#   - защищённая группа с RequireAuth для записи и пользовательских эндпоинтов
git add . && git commit -m "merge: единый router с защищёнными группами"

# 6. Сливаем frontend
# В App.vue и router/index.ts должны быть все основные экраны (карта, место,
# добавление места, вход, регистрация, подтверждение email, профиль, избранное).
# stores/places.ts (Атаханов) и stores/auth.ts (Насрулаев) живут параллельно.
# services/api.ts объединяется: prefix /api общий, токен подкладывается из stores/auth.
git add . && git commit -m "merge: единый SPA (все экраны)"

# 7. Расширяем агрегаты
# В services/places.go добавляется JOIN на reviews и AVG-агрегаты,
# Place в models.go получает поля avg_noise, avg_light, avg_crowd,
# avg_smell, avg_visual, overall_avg, reviews_count.
git add . && git commit -m "merge: интеграция агрегатов отзывов в карточку места"
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
| `backend/internal/models/Place`, `Category` | Атаханов Набиюлла Румиевич |
| `backend/internal/models/User`, `Review`, `Favorite`, `VerificationCode` | Насрулаев Шарапудин Махадович |
| `backend/internal/services/places.go` | Атаханов Набиюлла Румиевич |
| `backend/internal/services/users.go`, `verification.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/services/reviews.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/handlers/places_handler.go` | Атаханов Набиюлла Румиевич |
| `backend/internal/handlers/auth_handler.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/handlers/users_handler.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/handlers/reviews_handler.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/auth/jwt.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/middleware/auth.go` | Насрулаев Шарапудин Махадович |
| `backend/internal/database/database.go` (PostGIS, places) | Атаханов Набиюлла Румиевич |
| `backend/internal/database/database.go` (users, reviews seed) | Насрулаев Шарапудин Махадович |
| `backend/internal/routes/routes.go` | Атаханов Набиюлла Румиевич |
| `frontend/src/views/MapView.vue`, `AddPlaceView.vue`, `PlaceView.vue` | Атаханов Набиюлла Румиевич |
| `frontend/src/views/LoginView.vue`, `RegisterView.vue`, `ProfileView.vue`, `FavoritesView.vue`, `VerifyEmailView.vue` | Насрулаев Шарапудин Махадович |
| `frontend/src/components/FilterPanel.vue`, `PlaceCard.vue`, `SensoryRating.vue` | Атаханов Набиюлла Румиевич |
| `frontend/src/stores/places.ts` | Атаханов Набиюлла Румиевич |
| `frontend/src/stores/auth.ts` | Насрулаев Шарапудин Махадович |
| `frontend/src/services/api.ts` | Атаханов Набиюлла Румиевич |
| `frontend/src/router/index.ts`, `App.vue` | Атаханов Набиюлла Румиевич |
| `backend/docs/openapi.yaml` | Атаханов Набиюлла Румиевич |
| `docker-compose.yml` | Атаханов Набиюлла Румиевич |
| `backend/internal/email/sender.go` | Насрулаев Шарапудин Махадович |
