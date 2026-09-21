package hotel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hotelbooking/internal/hotel"
)

func TestParseStay(t *testing.T) {

	tests := []struct {
		name     string
		checkIn  string
		checkOut string
		want     string
		wantErr  bool
	}{
		{name: "one night", checkIn: "2026-10-01", checkOut: "2026-10-02", want: "2026-10-01/2026-10-02"},
		{name: "across a month end", checkIn: "2026-10-30", checkOut: "2026-11-02", want: "2026-10-30/2026-11-02"},
		{name: "check out before check in", checkIn: "2026-10-05", checkOut: "2026-10-01", wantErr: true},
		{name: "same day", checkIn: "2026-10-01", checkOut: "2026-10-01", wantErr: true},
		{name: "not a date", checkIn: "tomorrow", checkOut: "2026-10-01", wantErr: true},
		{name: "missing check out", checkIn: "2026-10-01", checkOut: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := hotel.ParseStay(tt.checkIn, tt.checkOut)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.CheckIn.Format("2006-01-02")+"/"+got.CheckOut.Format("2006-01-02"))
		})
	}
}
