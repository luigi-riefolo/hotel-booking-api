package booking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"hotelbooking/internal/hotel"
)

func TestValidate(t *testing.T) {

	now := time.Date(2026, 10, 1, 15, 0, 0, 0, time.UTC)

	stay := func(in, out string) hotel.Stay {

		st, err := hotel.ParseStay(in, out)
		if err != nil {
			t.Fatal(err)
		}

		return st
	}

	valid := Request{
		RoomID:     1,
		Guests:     2,
		GuestName:  "Homer Simpson",
		GuestEmail: "homer@springfield.com",
		Stay:       stay("2026-10-02", "2026-10-05"),
	}

	tests := []struct {
		name    string
		change  func(r *Request)
		wantErr bool
	}{
		{name: "valid request", change: func(r *Request) {}},
		{name: "checking in today", change: func(r *Request) { r.Stay = stay("2026-10-01", "2026-10-02") }},
		{name: "checking in yesterday", change: func(r *Request) { r.Stay = stay("2026-09-30", "2026-10-02") }, wantErr: true},
		{name: "no guests", change: func(r *Request) { r.Guests = 0 }, wantErr: true},
		{name: "blank name", change: func(r *Request) { r.GuestName = "  " }, wantErr: true},
		{name: "bad email", change: func(r *Request) { r.GuestEmail = "homer" }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := valid
			tt.change(&req)

			err := req.Validate(now)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestNewReference(t *testing.T) {

	ref := newReference()

	assert.Regexp(t, `^BK-[0-9A-F]{16}$`, ref)
	assert.NotEqual(t, ref, newReference())
}
