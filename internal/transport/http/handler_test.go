package http

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/auth"
	"github.com/Skuyno/room-booking/internal/domain"
	appmw "github.com/Skuyno/room-booking/internal/transport/http/middleware"
)

type authServiceStub struct {
	dummyLoginFn func(ctx context.Context, role string) (string, error)
}

func (s authServiceStub) DummyLogin(ctx context.Context, role string) (string, error) {
	return s.dummyLoginFn(ctx, role)
}

func (s authServiceStub) Register(ctx context.Context, email, password, role string) (domain.User, error) {
	return domain.User{}, nil
}

func (s authServiceStub) Login(ctx context.Context, email, password string) (string, error) {
	return "", nil
}

type roomServiceStub struct {
	createFn func(ctx context.Context, name string, description *string, capacity *int) (domain.Room, error)
}

func (s roomServiceStub) Create(ctx context.Context, name string, description *string, capacity *int) (domain.Room, error) {
	return s.createFn(ctx, name, description, capacity)
}

func (s roomServiceStub) List(ctx context.Context) ([]domain.Room, error) {
	return nil, nil
}

type scheduleServiceStub struct {
	createFn func(ctx context.Context, roomID uuid.UUID, daysOfWeek []int16, startTime, endTime string) (domain.Schedule, error)
}

func (s scheduleServiceStub) Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int16, startTime, endTime string) (domain.Schedule, error) {
	return s.createFn(ctx, roomID, daysOfWeek, startTime, endTime)
}

type slotServiceStub struct {
	listFn func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

func (s slotServiceStub) ListAvailableByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	return s.listFn(ctx, roomID, date)
}

type bookingServiceStub struct {
	createFn  func(ctx context.Context, userID uuid.UUID, role string, roomID uuid.UUID, startAt time.Time, createConferenceLink bool) (domain.Booking, error)
	cancelFn  func(ctx context.Context, userID uuid.UUID, role string, bookingID uuid.UUID) (domain.Booking, error)
	listAllFn func(ctx context.Context, role string, page, pageSize int) ([]domain.Booking, int, error)
}

func (s bookingServiceStub) Create(ctx context.Context, userID uuid.UUID, role string, roomID uuid.UUID, startAt time.Time, createConferenceLink bool) (domain.Booking, error) {
	return s.createFn(ctx, userID, role, roomID, startAt, createConferenceLink)
}

func (s bookingServiceStub) Cancel(ctx context.Context, userID uuid.UUID, role string, bookingID uuid.UUID) (domain.Booking, error) {
	return s.cancelFn(ctx, userID, role, bookingID)
}

func (s bookingServiceStub) ListMy(ctx context.Context, userID uuid.UUID, role string) ([]domain.Booking, error) {
	return nil, nil
}

func (s bookingServiceStub) ListAll(ctx context.Context, role string, page, pageSize int) ([]domain.Booking, int, error) {
	return s.listAllFn(ctx, role, page, pageSize)
}

func contextWithClaims(userID, role string) context.Context {
	claims := &auth.Claims{UserID: userID, Role: role}
	return context.WithValue(context.Background(), appmw.ClaimsKey, claims)
}

func TestActorFromContext(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	actor, err := actorFromContext(contextWithClaims(userID.String(), domain.RoleAdmin))
	if err != nil {
		t.Fatalf("actorFromContext() error = %v", err)
	}
	if actor.UserID != userID || actor.Role != domain.RoleAdmin {
		t.Fatalf("actor = %+v", actor)
	}

	_, err = actorFromContext(context.Background())
	if err == nil {
		t.Fatal("expected error without claims")
	}

	_, err = actorFromContext(contextWithClaims("bad-uuid", domain.RoleUser))
	if err == nil {
		t.Fatal("expected error for invalid uuid")
	}
}

func TestNormalizedName(t *testing.T) {
	t.Parallel()

	if normalizedName("  Omega  ") != "Omega" {
		t.Fatalf("normalizedName() = %q, want Omega", normalizedName("  Omega  "))
	}
}

func TestPostDummyLogin(t *testing.T) {
	t.Parallel()

	h := NewHandler(
		authServiceStub{
			dummyLoginFn: func(_ context.Context, role string) (string, error) {
				if role != domain.RoleUser {
					t.Fatalf("role = %s, want %s", role, domain.RoleUser)
				}
				return "token", nil
			},
		},
		roomServiceStub{},
		scheduleServiceStub{},
		slotServiceStub{},
		bookingServiceStub{},
	)

	resp, err := h.PostDummyLogin(context.Background(), api.PostDummyLoginRequestObject{})
	if err != nil {
		t.Fatalf("PostDummyLogin() error = %v", err)
	}
	if _, ok := resp.(api.PostDummyLogin400JSONResponse); !ok {
		t.Fatalf("response type = %T, want 400", resp)
	}

	body := api.PostDummyLoginJSONRequestBody{Role: api.PostDummyLoginJSONBodyRole(domain.RoleUser)}
	resp, err = h.PostDummyLogin(context.Background(), api.PostDummyLoginRequestObject{Body: &body})
	if err != nil {
		t.Fatalf("PostDummyLogin() error = %v", err)
	}
	if got, ok := resp.(api.PostDummyLogin200JSONResponse); !ok || got.Token != "token" {
		t.Fatalf("response = %#v, want token", resp)
	}
}

func TestPostRoomsCreateAuthAndSuccess(t *testing.T) {
	t.Parallel()

	description := "Main room"
	capacity := 6
	roomID := uuid.New()

	h := NewHandler(
		authServiceStub{},
		roomServiceStub{
			createFn: func(_ context.Context, name string, desc *string, cap *int) (domain.Room, error) {
				if name != "Omega" {
					t.Fatalf("name = %q, want Omega", name)
				}
				return domain.Room{ID: roomID, Name: name, Description: desc, Capacity: cap}, nil
			},
		},
		scheduleServiceStub{},
		slotServiceStub{},
		bookingServiceStub{},
	)

	body := api.PostRoomsCreateJSONRequestBody{Name: "  Omega  ", Description: &description, Capacity: &capacity}

	resp, err := h.PostRoomsCreate(context.Background(), api.PostRoomsCreateRequestObject{Body: &body})
	if err != nil {
		t.Fatalf("PostRoomsCreate() error = %v", err)
	}
	if _, ok := resp.(api.PostRoomsCreate401JSONResponse); !ok {
		t.Fatalf("response type = %T, want 401", resp)
	}

	resp, err = h.PostRoomsCreate(contextWithClaims(uuid.NewString(), domain.RoleUser), api.PostRoomsCreateRequestObject{Body: &body})
	if err != nil {
		t.Fatalf("PostRoomsCreate() error = %v", err)
	}
	if _, ok := resp.(api.PostRoomsCreate403JSONResponse); !ok {
		t.Fatalf("response type = %T, want 403", resp)
	}

	resp, err = h.PostRoomsCreate(contextWithClaims(uuid.NewString(), domain.RoleAdmin), api.PostRoomsCreateRequestObject{Body: &body})
	if err != nil {
		t.Fatalf("PostRoomsCreate() error = %v", err)
	}
	if got, ok := resp.(api.PostRoomsCreate201JSONResponse); !ok || got.Room == nil || got.Room.Id != roomID {
		t.Fatalf("response = %#v, want created room", resp)
	}
}

func TestPostRoomsRoomIdScheduleCreateSuccess(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	var gotDays []int16
	h := NewHandler(
		authServiceStub{},
		roomServiceStub{},
		scheduleServiceStub{
			createFn: func(_ context.Context, gotRoomID uuid.UUID, daysOfWeek []int16, startTime, endTime string) (domain.Schedule, error) {
				if gotRoomID != roomID {
					t.Fatalf("roomID = %s, want %s", gotRoomID, roomID)
				}
				gotDays = append([]int16(nil), daysOfWeek...)
				return domain.Schedule{ID: uuid.New(), RoomID: gotRoomID, DaysOfWeek: daysOfWeek, StartTime: startTime, EndTime: endTime}, nil
			},
		},
		slotServiceStub{},
		bookingServiceStub{},
	)

	body := api.PostRoomsRoomIdScheduleCreateJSONRequestBody{
		RoomId:     roomID,
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	resp, err := h.PostRoomsRoomIdScheduleCreate(
		contextWithClaims(uuid.NewString(), domain.RoleAdmin),
		api.PostRoomsRoomIdScheduleCreateRequestObject{RoomId: roomID, Body: &body},
	)
	if err != nil {
		t.Fatalf("PostRoomsRoomIdScheduleCreate() error = %v", err)
	}
	if len(gotDays) != 3 || gotDays[1] != 3 {
		t.Fatalf("gotDays = %v", gotDays)
	}
	if _, ok := resp.(api.PostRoomsRoomIdScheduleCreate201JSONResponse); !ok {
		t.Fatalf("response type = %T, want 201", resp)
	}
}

func TestGetRoomsRoomIdSlotsList(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	expectedDate := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	h := NewHandler(
		authServiceStub{},
		roomServiceStub{},
		scheduleServiceStub{},
		slotServiceStub{
			listFn: func(_ context.Context, gotRoomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
				if gotRoomID != roomID {
					t.Fatalf("roomID = %s, want %s", gotRoomID, roomID)
				}
				if !date.Equal(expectedDate) {
					t.Fatalf("date = %v, want %v", date, expectedDate)
				}
				return []domain.Slot{{
					RoomID:  roomID,
					StartAt: time.Date(2026, 4, 5, 9, 0, 0, 0, time.UTC),
					EndAt:   time.Date(2026, 4, 5, 9, 30, 0, 0, time.UTC),
				}}, nil
			},
		},
		bookingServiceStub{},
	)

	resp, err := h.GetRoomsRoomIdSlotsList(context.Background(), api.GetRoomsRoomIdSlotsListRequestObject{})
	if err != nil {
		t.Fatalf("GetRoomsRoomIdSlotsList() error = %v", err)
	}
	if _, ok := resp.(api.GetRoomsRoomIdSlotsList401JSONResponse); !ok {
		t.Fatalf("response type = %T, want 401", resp)
	}

	resp, err = h.GetRoomsRoomIdSlotsList(
		contextWithClaims(uuid.NewString(), domain.RoleUser),
		api.GetRoomsRoomIdSlotsListRequestObject{
			RoomId: roomID,
			Params: api.GetRoomsRoomIdSlotsListParams{
				Date: openapi_types.Date{Time: expectedDate},
			},
		},
	)
	if err != nil {
		t.Fatalf("GetRoomsRoomIdSlotsList() error = %v", err)
	}
	if _, ok := resp.(api.GetRoomsRoomIdSlotsList200JSONResponse); !ok {
		t.Fatalf("response type = %T, want 200", resp)
	}
}

func TestBookingsHandlers(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	roomID := uuid.New()
	bookingID := uuid.New()
	startAt := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)

	h := NewHandler(
		authServiceStub{},
		roomServiceStub{},
		scheduleServiceStub{},
		slotServiceStub{},
		bookingServiceStub{
			createFn: func(_ context.Context, gotUserID uuid.UUID, role string, gotRoomID uuid.UUID, gotStartAt time.Time, createConferenceLink bool) (domain.Booking, error) {
				if gotUserID != userID || role != domain.RoleUser || gotRoomID != roomID || !gotStartAt.Equal(startAt) || !createConferenceLink {
					t.Fatalf("unexpected create args: %s %s %s %v %v", gotUserID, role, gotRoomID, gotStartAt, createConferenceLink)
				}
				return domain.Booking{
					ID:      bookingID,
					RoomID:  roomID,
					UserID:  userID,
					Status:  domain.BookingStatusActive,
					StartAt: startAt,
					EndAt:   startAt.Add(30 * time.Minute),
				}, nil
			},
			listAllFn: func(_ context.Context, role string, page, pageSize int) ([]domain.Booking, int, error) {
				if role != domain.RoleAdmin || page != 1 || pageSize != 20 {
					t.Fatalf("unexpected list args: %s %d %d", role, page, pageSize)
				}
				return []domain.Booking{{
					ID:      bookingID,
					RoomID:  roomID,
					UserID:  userID,
					Status:  domain.BookingStatusActive,
					StartAt: startAt,
					EndAt:   startAt.Add(30 * time.Minute),
				}}, 1, nil
			},
			cancelFn: func(_ context.Context, gotUserID uuid.UUID, role string, gotBookingID uuid.UUID) (domain.Booking, error) {
				if gotUserID != userID || role != domain.RoleUser || gotBookingID != bookingID {
					t.Fatalf("unexpected cancel args")
				}
				return domain.Booking{
					ID:      bookingID,
					RoomID:  roomID,
					UserID:  userID,
					Status:  domain.BookingStatusCancelled,
					StartAt: startAt,
					EndAt:   startAt.Add(30 * time.Minute),
				}, nil
			},
		},
	)

	createLink := true
	createBody := api.PostBookingsCreateJSONRequestBody{
		RoomId:               roomID,
		StartAt:              startAt,
		CreateConferenceLink: &createLink,
	}
	createResp, err := h.PostBookingsCreate(
		contextWithClaims(userID.String(), domain.RoleUser),
		api.PostBookingsCreateRequestObject{Body: &createBody},
	)
	if err != nil {
		t.Fatalf("PostBookingsCreate() error = %v", err)
	}
	if _, ok := createResp.(api.PostBookingsCreate201JSONResponse); !ok {
		t.Fatalf("response type = %T, want 201", createResp)
	}

	listResp, err := h.GetBookingsList(
		contextWithClaims(userID.String(), domain.RoleAdmin),
		api.GetBookingsListRequestObject{},
	)
	if err != nil {
		t.Fatalf("GetBookingsList() error = %v", err)
	}
	if got, ok := listResp.(api.GetBookingsList200JSONResponse); !ok || got.Pagination == nil || got.Pagination.Total != 1 {
		t.Fatalf("response = %#v, want pagination", listResp)
	}

	cancelResp, err := h.PostBookingsBookingIdCancel(
		contextWithClaims(userID.String(), domain.RoleUser),
		api.PostBookingsBookingIdCancelRequestObject{BookingId: bookingID},
	)
	if err != nil {
		t.Fatalf("PostBookingsBookingIdCancel() error = %v", err)
	}
	if _, ok := cancelResp.(api.PostBookingsBookingIdCancel200JSONResponse); !ok {
		t.Fatalf("response type = %T, want 200", cancelResp)
	}
}
