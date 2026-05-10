package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Skuyno/room-booking/internal/domain"
)

// BookingRepository — pgx-реализация service.BookingRepository. Хранит
// брони как записи в таблице bookings со столбцами room_id/start_at/end_at
// (без отдельной таблицы slots).
type BookingRepository struct {
	q DBTX
}

// NewBookingRepository создаёт репозиторий поверх любого DBTX.
func NewBookingRepository(q DBTX) *BookingRepository {
	return &BookingRepository{
		q: q,
	}
}

// Create вставляет бронь. При попытке создать активную бронь на уже
// занятый (room_id, start_at) уникальный partial-индекс
// bookings_one_active_per_room_start_idx бросит ошибку — в этом случае
// возвращаем domain.ErrSlotAlreadyBooked.
func (r *BookingRepository) Create(ctx context.Context, booking domain.Booking) error {
	_, err := r.q.Exec(ctx, `
		insert into bookings (id, room_id, user_id, status_id, start_at, end_at, conference_link)
		values ($1, $2, $3, $4, $5, $6, $7)
	`,
		booking.ID,
		booking.RoomID,
		booking.UserID,
		booking.StatusID,
		booking.StartAt,
		booking.EndAt,
		booking.ConferenceLink,
	)
	if err != nil {
		if strings.Contains(err.Error(), "bookings_one_active_per_room_start_idx") {
			return domain.ErrSlotAlreadyBooked
		}
		return fmt.Errorf("insert booking: %w", err)
	}
	return nil
}

// GetByID возвращает бронь по ID или domain.ErrBookingNotFound.
func (r *BookingRepository) GetByID(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error) {
	booking, err := scanSingleBooking(ctx, r.q, `
		select id, room_id, user_id, status_id, start_at, end_at, conference_link, created_at, cancelled_at
		from bookings
		where id = $1
	`, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, domain.ErrBookingNotFound
		}
		return domain.Booking{}, fmt.Errorf("select booking by id: %w", err)
	}
	return booking, nil
}

// Cancel помечает бронь cancelled. cancelled_at выставляется через
// COALESCE: если бронь уже была отменена раньше, исходное время сохраняется.
// Это и обеспечивает идемпотентность.
func (r *BookingRepository) Cancel(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error) {
	booking, err := scanSingleBooking(ctx, r.q, `
		update bookings
		set status_id = $2,
		    cancelled_at = coalesce(cancelled_at, now())
		where id = $1
		returning id, room_id, user_id, status_id, start_at, end_at, conference_link, created_at, cancelled_at
	`, bookingID, domain.BookingStatusIDCancelled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return domain.Booking{}, domain.ErrBookingNotFound
		}
		return domain.Booking{}, fmt.Errorf("cancel booking: %w", err)
	}
	return booking, nil
}

// ListByUserFuture возвращает только будущие брони пользователя (start_at >= now),
// отсортированные по возрастанию start_at.
func (r *BookingRepository) ListByUserFuture(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) ([]domain.Booking, error) {
	rows, err := r.q.Query(ctx, `
		select id, room_id, user_id, status_id, start_at, end_at, conference_link, created_at, cancelled_at
		from bookings
		where user_id = $1
		  and start_at >= $2
		order by start_at asc
	`, userID, now.UTC())
	if err != nil {
		return nil, fmt.Errorf("list my bookings: %w", err)
	}
	defer rows.Close()

	return scanBookingRows(rows)
}

// ListAll возвращает страницу всех броней и общее количество. Делает два
// запроса: COUNT(*) и SELECT с LIMIT/OFFSET. Без транзакции total и items
// могут немного разъехаться при параллельной записи — для admin-листинга
// это допустимо.
func (r *BookingRepository) ListAll(
	ctx context.Context,
	page, pageSize int,
) ([]domain.Booking, int, error) {
	var total int
	if err := r.q.QueryRow(ctx, `select count(*) from bookings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookings: %w", err)
	}

	offset := (page - 1) * pageSize

	rows, err := r.q.Query(ctx, `
		select id, room_id, user_id, status_id, start_at, end_at, conference_link, created_at, cancelled_at
		from bookings
		order by created_at desc
		limit $1 offset $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	bookings, err := scanBookingRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}

// ListActiveStartsByRoomBetween возвращает start_at активных броней комнаты
// в полуоткрытом интервале [from, to). Используется SlotService для
// фильтрации занятых слотов из выдачи.
func (r *BookingRepository) ListActiveStartsByRoomBetween(
	ctx context.Context,
	roomID uuid.UUID,
	from, to time.Time,
) ([]time.Time, error) {
	rows, err := r.q.Query(ctx, `
		select start_at
		from bookings
		where room_id = $1
		  and status_id = $2
		  and start_at >= $3
		  and start_at < $4
	`, roomID, domain.BookingStatusIDActive, from.UTC(), to.UTC())
	if err != nil {
		return nil, fmt.Errorf("list active bookings: %w", err)
	}
	defer rows.Close()

	var result []time.Time
	for rows.Next() {
		var startAt time.Time
		if err = rows.Scan(&startAt); err != nil {
			return nil, fmt.Errorf("scan active booking: %w", err)
		}
		result = append(result, startAt.UTC())
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active bookings: %w", err)
	}

	return result, nil
}

// scanSingleBooking выполняет QueryRow и складывает результат в domain.Booking.
// NULL-поля (conference_link, cancelled_at) проходят через sql.Null*-обёртки.
// Сделан общим, чтобы GetByID и Cancel не дублировали этот код.
func scanSingleBooking(ctx context.Context, q DBTX, query string, args ...any) (domain.Booking, error) {
	var booking domain.Booking
	var conferenceLink sql.NullString
	var cancelledAt sql.NullTime

	err := q.QueryRow(ctx, query, args...).Scan(
		&booking.ID,
		&booking.RoomID,
		&booking.UserID,
		&booking.StatusID,
		&booking.StartAt,
		&booking.EndAt,
		&conferenceLink,
		&booking.CreatedAt,
		&cancelledAt,
	)
	if err != nil {
		return domain.Booking{}, err
	}

	booking.Status = domain.BookingStatusCodeByID(booking.StatusID)
	if conferenceLink.Valid {
		booking.ConferenceLink = &conferenceLink.String
	}
	if cancelledAt.Valid {
		booking.CancelledAt = &cancelledAt.Time
	}

	return booking, nil
}

// scanBookingRows скан-ит результат Query() в []domain.Booking.
// Аналогично scanSingleBooking, но для множественной выдачи — DRY между
// ListByUserFuture и ListAll.
func scanBookingRows(rows pgx.Rows) ([]domain.Booking, error) {
	var result []domain.Booking
	for rows.Next() {
		var booking domain.Booking
		var conferenceLink sql.NullString
		var cancelledAt sql.NullTime

		if err := rows.Scan(
			&booking.ID,
			&booking.RoomID,
			&booking.UserID,
			&booking.StatusID,
			&booking.StartAt,
			&booking.EndAt,
			&conferenceLink,
			&booking.CreatedAt,
			&cancelledAt,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		booking.Status = domain.BookingStatusCodeByID(booking.StatusID)
		if conferenceLink.Valid {
			booking.ConferenceLink = &conferenceLink.String
		}
		if cancelledAt.Valid {
			booking.CancelledAt = &cancelledAt.Time
		}

		result = append(result, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return result, nil
}
