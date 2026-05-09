package http

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/domain"
	appmw "github.com/Skuyno/room-booking/internal/transport/http/middleware"
)

// AuthService — контракт, который Handler ожидает от сервиса аутентификации.
// Объявлен здесь, а не импортируется из service, чтобы транспорт не зависел
// от конкретной реализации (можно подменить при тестировании).
type AuthService interface {
	DummyLogin(ctx context.Context, role string) (string, error)
	Register(ctx context.Context, email, password, role string) (domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// RoomService — контракт сервиса переговорок для HTTP-слоя.
type RoomService interface {
	Create(ctx context.Context, name string, description *string, capacity *int) (domain.Room, error)
	List(ctx context.Context) ([]domain.Room, error)
}

// ScheduleService — контракт сервиса расписаний для HTTP-слоя.
type ScheduleService interface {
	Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int16, startTime, endTime string) (domain.Schedule, error)
}

// SlotService — контракт сервиса свободных слотов для HTTP-слоя.
type SlotService interface {
	ListAvailableByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

// BookingService — контракт сервиса броней для HTTP-слоя.
type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, role string, roomID uuid.UUID, startAt time.Time, createConferenceLink bool) (domain.Booking, error)
	Cancel(ctx context.Context, userID uuid.UUID, role string, bookingID uuid.UUID) (domain.Booking, error)
	ListMy(ctx context.Context, userID uuid.UUID, role string) ([]domain.Booking, error)
	ListAll(ctx context.Context, role string, page, pageSize int) ([]domain.Booking, int, error)
}

// Handler реализует api.StrictServerInterface — все HTTP-ручки приложения.
// Хранит ссылки на сервисы и делегирует им бизнес-логику; сам только
// разбирает ввод, маппит ошибки и формирует ответы.
type Handler struct {
	authService     AuthService
	roomService     RoomService
	scheduleService ScheduleService
	slotService     SlotService
	bookingService  BookingService
}

// NewHandler собирает Handler из набора сервисов.
func NewHandler(
	authService AuthService,
	roomService RoomService,
	scheduleService ScheduleService,
	slotService SlotService,
	bookingService BookingService,
) *Handler {
	return &Handler{
		authService:     authService,
		roomService:     roomService,
		scheduleService: scheduleService,
		slotService:     slotService,
		bookingService:  bookingService,
	}
}

// Compile-time проверка, что Handler полностью реализует
// сгенерированный StrictServerInterface — пропуск метода ловится
// при компиляции, а не в рантайме.
var _ api.StrictServerInterface = (*Handler)(nil)

// actor — данные авторизованного вызова, извлечённые из JWT-claims.
type actor struct {
	UserID uuid.UUID
	Role   string
}

// actorFromContext достаёт claims, кладущиеся middleware/auth.go в контекст,
// и парсит UserID. Возвращает ошибку, если claims нет (запрос не прошёл
// auth middleware) или UserID не uuid. Хендлеры превращают такую ошибку
// в 401 UNAUTHORIZED.
func actorFromContext(ctx context.Context) (actor, error) {
	claims, ok := appmw.GetClaims(ctx)
	if !ok {
		return actor{}, fmt.Errorf("claims not found in context")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return actor{}, fmt.Errorf("parse user id from token: %w", err)
	}

	return actor{
		UserID: userID,
		Role:   claims.Role,
	}, nil
}

// normalizedName убирает ведущие/замыкающие пробелы. Используется для
// строковых имён комнат, чтобы хранить и сравнивать без артефактов.
func normalizedName(value string) string {
	return strings.TrimSpace(value)
}
