package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

func TestRoomServiceCreateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	svc := NewRoomService(roomRepoStub{})
	capacity := 0

	_, err := svc.Create(context.Background(), "", nil, nil)
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("Create() empty name error = %v, want %v", err, domain.ErrInvalidRequest)
	}

	_, err = svc.Create(context.Background(), "Room", nil, &capacity)
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("Create() invalid capacity error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestRoomServiceCreatePersistsRoom(t *testing.T) {
	t.Parallel()

	description := "Meeting room"
	capacity := 6
	var saved domain.Room

	repo := roomRepoStub{
		createFn: func(_ context.Context, room domain.Room) error {
			saved = room
			return nil
		},
	}

	svc := NewRoomService(repo)

	room, err := svc.Create(context.Background(), "Omega", &description, &capacity)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if saved.Name != "Omega" {
		t.Fatalf("saved.Name = %s, want Omega", saved.Name)
	}
	if saved.Description == nil || *saved.Description != description {
		t.Fatalf("saved.Description = %v, want %s", saved.Description, description)
	}
	if saved.Capacity == nil || *saved.Capacity != capacity {
		t.Fatalf("saved.Capacity = %v, want %d", saved.Capacity, capacity)
	}
	if room.ID == uuid.Nil {
		t.Fatal("room.ID must be generated")
	}
}
