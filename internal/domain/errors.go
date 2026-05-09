// Package domain содержит ядро предметной области: сущности (комнаты,
// расписания, брони, пользователи), их инварианты и типизированные ошибки.
// Пакет не зависит ни от БД, ни от HTTP, ни от JSON — только стандартная
// библиотека и uuid. Любой слой выше (service, transport) опирается на
// domain, но не наоборот.
package domain

import "errors"

// Доменные ошибки. Слой service возвращает их через fmt.Errorf("...: %w", ...)
// или напрямую, а HTTP-слой превращает в коды ответа через mapDomainError.
// Сравнение ВСЕГДА через errors.Is, чтобы wrap не ломал маппинг.
var (
	// ErrRoomNotFound — переговорки с указанным ID не существует.
	ErrRoomNotFound = errors.New("room not found")
	// ErrScheduleExists — для комнаты уже создано расписание; по дизайну
	// расписание создаётся один раз и не меняется.
	ErrScheduleExists = errors.New("schedule already exists")
	// ErrScheduleNotFound — у комнаты ещё нет расписания. SlotService
	// трактует это как «свободных слотов нет», BookingService — как
	// «забронировать здесь нечего».
	ErrScheduleNotFound = errors.New("schedule not found")
	// ErrSlotNotFound — нет валидного слота на запрошенное (room, startAt):
	// либо время вне расписания, либо день недели не разрешён, либо
	// startAt не лежит на 30-минутной сетке.
	ErrSlotNotFound = errors.New("slot not found")
	// ErrSlotAlreadyBooked — пара (room_id, start_at) уже занята активной
	// бронью. Возвращается, когда уникальный partial-индекс
	// bookings_one_active_per_room_start_idx ловит конфликт.
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	// ErrBookingNotFound — брони с указанным ID нет.
	ErrBookingNotFound = errors.New("booking not found")
	// ErrForbidden — действие запрещено для текущей роли (например, admin
	// пытается создать бронь, или user пытается создать комнату).
	ErrForbidden = errors.New("forbidden")
	// ErrInvalidRequest — общая ошибка валидации входа: формат, диапазоны,
	// прошлое время, нарушение бизнес-правил формата.
	ErrInvalidRequest = errors.New("invalid request")
	// ErrEmailTaken — попытка зарегистрировать пользователя с занятым email.
	ErrEmailTaken = errors.New("email already taken")
	// ErrInvalidCredentials — комбинация email+пароль не подходит. Один и
	// тот же код для «нет такого юзера» и «пароль не верен», чтобы не
	// сигналить наружу о существовании email.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound — пользователь не найден; внутренняя ошибка
	// репозитория, наружу не выдаётся (Login превращает в ErrInvalidCredentials).
	ErrUserNotFound = errors.New("user not found")
)
