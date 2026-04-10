# Wishlist API

REST API сервис на Go для управления вишлистами:
- регистрация и логин по email/паролю;
- CRUD вишлистов и подарков;
- публичный доступ к вишлисту по токен-ссылке;
- бронирование подарка без авторизации.

## Грустная истоория (обязательно к прочтению)

PS: В процессе разработки я часто делал коммиты, разбивая задачи на логические этапы. Однако в конце работы, при попытке переключиться на другую ветку локально для финальной полировки, я по ошибке выполнил git reset --hard не на той ветке (это было грустно очень), что привело к полной потере всех незапушенных коммитов (а это были почти все коммиты) и сбросу рабочей директории до начального состояния.

К счастью, у меня оставались локальные бекапы исходного кода (я периодически архивировал папку проекта). Я восстановил актуальное состояние кода, но история коммитов была утеряна. Чтобы сохранить читаемую структуру изменений, я вручную воссоздал последовательность коммитов, повторив их в том же логическом порядке, в котором они делались изначально.

## Что реализовано по ТЗ

- Авторизация: `register/login`, пароли хранятся в bcrypt-хэше, закрытые роуты защищены JWT.
- Вишлисты: полный CRUD только для владельца.
- Позиции: полный CRUD внутри вишлиста владельца.
- Публичные эндпоинты:
  - получение вишлиста с позициями по `public_token`;
  - бронирование подарка (повторная бронь возвращает `409 Conflict`).
- Валидация входных данных и JSON-ошибки с корректными HTTP-кодами.
- SQL-миграции через `golang-migrate`.
- Docker-инфраструктура для запуска одной командой.
- Unit-тесты для бизнес-логики и валидации.

## Технологии

- Go 1.22
- PostgreSQL 16
- Docker Compose
- `chi` (роутинг), `pgx` (БД), `golang-migrate`, `jwt/v5`, `bcrypt`

## Быстрый старт (как в задании)

```bash
docker-compose up --build
```

Этого достаточно для старта сервиса.  
API будет доступен на `http://localhost:8080`.

Если хотите переопределить конфигурацию, создайте `.env`:

```bash
cp .env.example .env
docker-compose up --build
```

## Конфигурация (.env / .env.example)

| Переменная | Назначение |
|---|---|
| `HTTP_ADDR` | Адрес API внутри контейнера (по умолчанию `:8080`) |
| `POSTGRES_USER` | Пользователь PostgreSQL |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL |
| `POSTGRES_DB` | Имя базы данных |
| `POSTGRES_PORT` | Порт PostgreSQL на хосте |
| `DATABASE_URL` | DSN подключения API к PostgreSQL |
| `JWT_SECRET` | Секрет для подписи JWT (в проде замените на надежный) |
| `JWT_EXPIRATION_HOURS` | TTL токена в часах |
| `MIGRATIONS_PATH` | Путь к SQL-миграциям в контейнере |

## API

Все ответы в JSON.  
Формат ошибки: `{"error":"..."}`.

### 1) Авторизация

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

Тело запроса:

```json
{
  "email": "alice@example.com",
  "password": "secretpass123"
}
```

В ответе: пользователь и JWT-токен.

Для защищенных эндпоинтов используйте:

`Authorization: Bearer <token>`

### 2) Вишлисты (только свои)

- `POST /api/v1/wishlists`
- `GET /api/v1/wishlists`
- `GET /api/v1/wishlists/{id}`
- `PATCH /api/v1/wishlists/{id}`
- `DELETE /api/v1/wishlists/{id}`

Пример создания:

```json
{
  "title": "Birthday",
  "description": "Ideas for gifts",
  "event_date": "2026-12-31"
}
```

При создании автоматически генерируется `public_token`.

### 3) Позиции в вишлисте

- `POST /api/v1/wishlists/{wishlistID}/items`
- `GET /api/v1/wishlists/{wishlistID}/items`
- `GET /api/v1/wishlists/{wishlistID}/items/{itemID}`
- `PATCH /api/v1/wishlists/{wishlistID}/items/{itemID}`
- `DELETE /api/v1/wishlists/{wishlistID}/items/{itemID}`

Пример создания позиции:

```json
{
  "title": "Go Book",
  "description": "Advanced topics",
  "product_url": "https://example.com/go-book",
  "priority": 9
}
```

### 4) Публичный доступ по ссылке

- `GET /public/v1/wishlists/{token}` — получить вишлист с позициями без авторизации
- `POST /public/v1/wishlists/{token}/items/{itemID}/reserve` — забронировать подарок без авторизации

Если подарок уже забронирован: `409 Conflict`.

### 5) Healthcheck

- `GET /health` -> `200 OK`, тело: `ok`

## Примеры запросов (curl)

```bash
# Регистрация
curl -sS -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secretpass123"}'

# Логин
curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secretpass123"}'
```

## Тесты

```bash
go test ./...
```

Покрыты:
- `internal/logic`
- `internal/validation`

## Замечания для проверки

- Миграции применяются автоматически при старте API.
- Секреты не должны храниться в git: файл `.env` добавлен в `.gitignore`.
- Для полностью "чистого" старта Docker-окружения можно удалить старые тома:

```bash
docker-compose down -v
docker-compose up --build
```
