# Room Booking Service

Учебный backend-сервис бронирования переговорных комнат на Go + PostgreSQL.

OpenAPI-спека: [api/api.yaml](./api/api.yaml).

## Возможности

- регистрация / логин (JWT, bcrypt) + dev-only `dummyLogin` под фиксированные UUID;
- админ создаёт переговорки и расписание (дни недели + окно `HH:MM`);
- слоты по 30 минут вычисляются на лету из расписания и не материализуются в БД;
- пользователь бронирует слот по `(roomId, startAt)`, отменяет свою бронь (идемпотентно);
- админ-листинг броней с пагинацией;
- автоприменение SQL-миграций при старте.

## Стек

Go 1.25 · chi · pgx/v5 · oapi-codegen (strict-server) · PostgreSQL 16 · Docker Compose.

## Архитектура

```
cmd/api                    composition root
└── internal/transport/http handlers, mappers, middleware
    └── internal/service    бизнес-логика (зависит от интерфейсов репозиториев)
        ├── internal/storage/postgres  pgx-репозитории
        └── internal/domain  сущности и доменные ошибки
```

Сервисы зависят только от интерфейсов репозиториев — конкретный PostgreSQL
подключается в [cmd/api/main.go](./cmd/api/main.go).

### Ключевые решения

- **Slots не хранятся в БД** — вычисляются из расписания. Уникальность
  активной брони — partial unique index `(room_id, start_at) WHERE status = active`.
- **Расписание неизменяемо** после создания (constraint `schedules_room_unique`).
- **30-минутная сетка** зашита и в коде (`SlotDuration`), и в check-constraint
  `bookings_duration_30m` — нельзя обойти на уровне БД.
- **UTC везде**: на входе, в БД, на выходе.

## Запуск

### Docker Compose (рекомендуется)

```bash
cp .env.example .env  # отредактируй JWT_SECRET
docker compose up --build
```

API на `http://localhost:8080`, health-check на `/_info`.

### Локально

Требуется PostgreSQL 16+. Переменные окружения:

| | |
|---|---|
| `HTTP_PORT` | порт HTTP-сервера |
| `DB_DSN` | DSN PostgreSQL |
| `JWT_SECRET` | симметричный ключ для подписи JWT |

```bash
go run ./cmd/api
```

## Тесты

```bash
go test ./...                       # все
go test ./internal/...              # только unit
TEST_DB_DSN=... go test ./tests/integration/...   # integration
```

Integration-тесты используют отдельную БД через `TEST_DB_DSN`; БД очищается
через `TRUNCATE`, миграции запускаются автоматически.

## API кратко

| Метод | Путь | Кто |
|---|---|---|
| POST | `/dummyLogin` | dev-only |
| POST | `/register`, `/login` | публично |
| GET | `/_info` | публично |
| GET, POST | `/rooms/list`, `/rooms/create` | auth, create — admin |
| POST | `/rooms/{roomId}/schedule/create` | admin |
| GET | `/rooms/{roomId}/slots/list?date=YYYY-MM-DD` | auth |
| POST | `/bookings/create` | user |
| GET | `/bookings/my` | user |
| GET | `/bookings/list` | admin |
| POST | `/bookings/{bookingId}/cancel` | user (только своя) |

## Лицензия

MIT — см. [LICENSE](./LICENSE).
