# Сенсорный навигатор

Карта с отзывами для людей, склонных к сенсорной перегрузке.

Кроссплатформенный десктопный сервис для людей с расстройством аутистического
спектра (РАС) и повышенной сенсорной чувствительностью. Позволяет планировать
посещения общественных мест с учётом уровня шума, освещённости, заполненности
и других сенсорных характеристик. В отличие от Google Maps, 2ГИС и Яндекс
Карт, сервис фокусируется не на качестве услуг, а на комфорте сенсорной среды.

> Курсовой проект, ФКН НИУ ВШЭ, ОП «Программная инженерия», 2025/2026.
>
> Команда:
> - **Атаханов Набиюлла Румиевич** (БПИ234) — модуль карты, мест, геопоиска, фильтрации
>   (`backend/internal/services/places.go`,
>   `backend/internal/handlers/places_handler.go`,
>   фронтенд: `frontend/src/views/MapView.vue`,
>   `frontend/src/components/{FilterPanel,PlaceCard,SensoryRating}.vue`).
> - **Насрулаев Шарапудин Махадович** (БПИ234) — модуль пользователей, авторизации, отзывов,
>   профилей (`backend/internal/services/{users,reviews}.go`,
>   `backend/internal/handlers/{auth,users,reviews}_handler.go`,
>   фронтенд: `frontend/src/views/{Login,Register,Profile,Place,Favorites}View.vue`).
>
> Научный руководитель — А. К. Бегичева, ст. преп. департамента программной
> инженерии.

## Технологический стек

### Backend
- **Язык:** Go 1.21+
- **Фреймворк:** Gin (`github.com/gin-gonic/gin`)
- **ORM:** GORM (`gorm.io/gorm`)
- **СУБД:** PostgreSQL 16 + PostGIS 3.4
- **Авторизация:** JWT (`github.com/golang-jwt/jwt/v5`), bcrypt
- **API:** REST + JSON, спецификация OpenAPI 3.0 (`backend/docs/openapi.yaml`)
- **Миграции:** автоматические через GORM, инициализация PostGIS-колонки
  `location` и индекса GIST.

### Frontend
- **Фреймворк:** Tauri 2.0 (Rust + WebView)
- **UI-фреймворк:** Vue 3 + Pinia + Vue Router
- **Сборка:** Vite + TypeScript
- **Карта:** Leaflet 1.9 + OpenStreetMap-тайлы + Leaflet.markercluster

Фронтенд работает и как обычное web-приложение (`npm run dev`), и как нативное
десктопное приложение (`npm run tauri:dev`) под Windows / macOS / Linux.

## Архитектура

```
sensory-navigator/
├── backend/
│   ├── cmd/
│   │   ├── server/         # точка входа сервера
│   │   └── seed/           # утилита заполнения БД демо-данными
│   ├── internal/
│   │   ├── auth/           # JWT + bcrypt
│   │   ├── config/         # переменные окружения
│   │   ├── database/       # подключение, миграции, сидеры
│   │   ├── handlers/       # HTTP-обработчики
│   │   ├── middleware/     # auth-middleware
│   │   ├── models/         # модели данных (GORM)
│   │   ├── routes/         # сборка роутера
│   │   └── services/       # бизнес-логика
│   ├── docs/openapi.yaml   # спецификация API
│   ├── tests/              # интеграционные тесты
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/     # переиспользуемые компоненты
│   │   ├── router/         # Vue Router
│   │   ├── services/       # api-клиент
│   │   ├── stores/         # Pinia-сторы
│   │   ├── types/          # TS-типы
│   │   ├── views/          # экраны
│   │   ├── App.vue
│   │   └── main.ts
│   ├── src-tauri/          # Rust-обёртка Tauri
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
└── docker-compose.yml      # запускает PostgreSQL+PostGIS и backend
```

## Быстрый старт

### 1. Запуск backend и БД через Docker

```bash
cd Project
docker compose up -d --build
```

Сервер становится доступен по адресу `http://localhost:8080`. PostgreSQL —
на `localhost:5432` (`navigator/navigator`, БД `sensory_navigator`).

Для загрузки демо-данных (демо-пользователь `demo@example.com / demo123` и
несколько мест Москвы):

```bash
docker compose exec backend /app/server &  # уже работает
docker compose run --rm --entrypoint "" backend sh -c "go run ./cmd/seed"
```

(или локально: `cd backend && go run ./cmd/seed`)

### 2. Запуск frontend

#### а) Web-режим (для разработки и быстрого тестирования)

```bash
cd frontend
npm install
npm run dev
```

Открыть `http://localhost:1420`.

#### б) Десктопный режим (Tauri 2.0)

Требуется установленный Rust (`rustup`). Подробности —
<https://tauri.app/start/prerequisites/>.

```bash
cd frontend
npm install
npm run tauri:dev          # запуск
npm run tauri:build        # сборка нативного исполняемого файла
```

После успешной сборки исполняемый файл появится в
`frontend/src-tauri/target/release/bundle/`.

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `APP_PORT` | `8080` | порт HTTP-сервера |
| `DB_HOST` | `localhost` | хост PostgreSQL |
| `DB_PORT` | `5432` | порт PostgreSQL |
| `DB_USER` | `navigator` | пользователь PostgreSQL |
| `DB_PASSWORD` | `navigator` | пароль PostgreSQL |
| `DB_NAME` | `sensory_navigator` | имя БД |
| `JWT_SECRET` | — | секрет для подписи JWT (обязательно сменить в проде) |
| `JWT_TTL` | `168h` | время жизни токена |
| `BCRYPT_COST` | `10` | сложность bcrypt |
| `VITE_API_BASE` | `http://localhost:8080` | адрес API для фронтенда |

## Тестирование

### Backend

```bash
cd backend
go test ./...
```

Юнит-тесты `internal/auth/...` и `internal/services/...` запускаются без
внешних зависимостей. Интеграционные HTTP-тесты в `tests/...` активируются
только при наличии переменной `TEST_DB_DSN` (DSN PostgreSQL).

Пример с локальной тестовой БД:

```bash
TEST_DB_DSN="host=localhost port=5432 user=navigator password=navigator \
dbname=sensory_navigator_test sslmode=disable" go test ./tests/...
```

### Frontend

Ручные сценарии описаны в документе
`Документация/Готовые/ПМИ_Атаханов.pdf`.

## API

Полная спецификация — в `backend/docs/openapi.yaml`. Краткий список:

| Метод | Путь | Auth | Назначение |
|-------|------|------|------------|
| `POST` | `/api/auth/register` | — | регистрация |
| `POST` | `/api/auth/login` | — | вход |
| `GET` | `/api/categories` | — | категории мест |
| `GET` | `/api/places` | — | места + фильтры |
| `GET` | `/api/places/search?q=` | — | поиск по названию/адресу |
| `GET` | `/api/places/nearby?lat=&lon=&radius=` | — | геопоиск (PostGIS / Хаверсин) |
| `GET` | `/api/places/:id` | — | карточка места |
| `POST` | `/api/places` | JWT | создать место |
| `PUT` | `/api/places/:id` | JWT | изменить |
| `DELETE` | `/api/places/:id` | JWT | удалить |
| `GET` | `/api/places/:id/reviews` | — | отзывы о месте |
| `POST` | `/api/places/:id/reviews` | JWT | оставить отзыв |
| `PUT` | `/api/reviews/:id` | JWT | изменить свой |
| `DELETE` | `/api/reviews/:id` | JWT | удалить свой |
| `GET` | `/api/reviews/me` | JWT | мои отзывы |
| `GET` | `/api/me` / `PUT` / `PUT /password` | JWT | профиль |
| `POST` / `DELETE` | `/api/places/:id/favorite` | JWT | избранное |
| `GET` | `/api/favorites` | JWT | список избранного |

## Лицензия

MIT.
