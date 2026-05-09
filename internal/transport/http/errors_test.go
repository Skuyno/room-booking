package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/domain"
)

func TestMapDomainError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   api.ErrorResponseErrorCode
	}{
		{name: "invalid request", err: domain.ErrInvalidRequest, wantStatus: stdhttp.StatusBadRequest, wantCode: api.INVALIDREQUEST},
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: stdhttp.StatusForbidden, wantCode: api.FORBIDDEN},
		{name: "room not found", err: domain.ErrRoomNotFound, wantStatus: stdhttp.StatusNotFound, wantCode: api.ROOMNOTFOUND},
		{name: "slot already booked", err: domain.ErrSlotAlreadyBooked, wantStatus: stdhttp.StatusConflict, wantCode: api.SLOTALREADYBOOKED},
		{name: "default", err: errors.New("boom"), wantStatus: stdhttp.StatusInternalServerError, wantCode: api.INTERNALERROR},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := mapDomainError(tc.err)
			if got.Status != tc.wantStatus {
				t.Fatalf("Status = %d, want %d", got.Status, tc.wantStatus)
			}
			if got.Code != tc.wantCode {
				t.Fatalf("Code = %s, want %s", got.Code, tc.wantCode)
			}
		})
	}
}

func TestStrictErrorHandlersReturnJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(stdhttp.MethodGet, "/", nil)

	t.Run("request error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleStrictRequestError(rec, req, errors.New("bad request"))

		if rec.Code != stdhttp.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %s, want application/json", ct)
		}

		var body api.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if body.Error.Code != api.INVALIDREQUEST {
			t.Fatalf("body.Error.Code = %s, want %s", body.Error.Code, api.INVALIDREQUEST)
		}
	})

	t.Run("response error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handleStrictResponseError(rec, req, errors.New("internal"))

		if rec.Code != stdhttp.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}

		var body api.InternalErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if body.Error.Code != string(api.INTERNALERROR) {
			t.Fatalf("body.Error.Code = %s, want %s", body.Error.Code, api.INTERNALERROR)
		}
	})
}
