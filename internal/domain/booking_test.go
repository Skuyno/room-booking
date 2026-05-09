package domain

import "testing"

func TestBookingStatusMappings(t *testing.T) {
	t.Parallel()

	if BookingStatusCodeByID(BookingStatusIDActive) != BookingStatusActive {
		t.Fatalf("BookingStatusCodeByID(active) = %q", BookingStatusCodeByID(BookingStatusIDActive))
	}
	if BookingStatusCodeByID(BookingStatusIDCancelled) != BookingStatusCancelled {
		t.Fatalf("BookingStatusCodeByID(cancelled) = %q", BookingStatusCodeByID(BookingStatusIDCancelled))
	}
	if BookingStatusCodeByID(99) != "" {
		t.Fatalf("BookingStatusCodeByID(99) = %q, want empty string", BookingStatusCodeByID(99))
	}

	if BookingStatusIDByCode(BookingStatusActive) != BookingStatusIDActive {
		t.Fatalf("BookingStatusIDByCode(active) = %d", BookingStatusIDByCode(BookingStatusActive))
	}
	if BookingStatusIDByCode(BookingStatusCancelled) != BookingStatusIDCancelled {
		t.Fatalf("BookingStatusIDByCode(cancelled) = %d", BookingStatusIDByCode(BookingStatusCancelled))
	}
	if BookingStatusIDByCode("other") != 0 {
		t.Fatalf("BookingStatusIDByCode(other) = %d, want 0", BookingStatusIDByCode("other"))
	}
}
