// Command api — точка входа сервиса бронирования переговорок.
//
// Что делает main:
//  1. Загружает конфиг из переменных окружения.
//  2. Открывает пул соединений к PostgreSQL и применяет миграции.
//  3. Собирает репозитории, сервисы и HTTP-обработчик (composition root).
//  4. Запускает HTTP-сервер с graceful shutdown по SIGINT/SIGTERM.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skuyno/room-booking/internal/auth"
	"github.com/Skuyno/room-booking/internal/config"
	"github.com/Skuyno/room-booking/internal/service"
	"github.com/Skuyno/room-booking/internal/storage/postgres"
	apphttp "github.com/Skuyno/room-booking/internal/transport/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := postgres.NewPool(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(ctx, db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Репозитории объявлены через интерфейсы service.* — конкретные
	// postgres-реализации видны только тут, в composition root.
	var roomRepo service.RoomRepository = postgres.NewRoomRepository(db)
	var scheduleRepo service.ScheduleRepository = postgres.NewScheduleRepository(db)
	var bookingRepo service.BookingRepository = postgres.NewBookingRepository(db)
	var userRepo service.UserRepository = postgres.NewUserRepository(db)

	authService := service.NewAuthService(jwtManager, userRepo)
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo)
	slotService := service.NewSlotService(roomRepo, scheduleRepo, bookingRepo)
	bookingService := service.NewBookingService(scheduleRepo, bookingRepo)

	handler := apphttp.NewHandler(
		authService,
		roomService,
		scheduleService,
		slotService,
		bookingService,
	)

	router := apphttp.NewServer(handler, jwtManager)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("http server started on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	// Блокируемся до сигнала остановки. signal.NotifyContext отменит ctx
	// при получении SIGINT/SIGTERM.
	<-ctx.Done()

	// Даём серверу 10 секунд на завершение активных запросов.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
