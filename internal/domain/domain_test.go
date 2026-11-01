package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	d "github.com/SIniutin/room-booking-service/internal/domain"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"valid lowercase", "user@example.com", "user@example.com", nil},
		{"normalises to lower", "User@Example.COM", "user@example.com", nil},
		{"trims spaces", "  user@example.com  ", "user@example.com", nil},
		{"empty", "", "", d.ErrEmptyEmail},
		{"only spaces", "   ", "", d.ErrEmptyEmail},
		{"no at-sign", "notanemail", "", d.ErrInvalidEmailFormat},
		{"no domain", "user@", "", d.ErrInvalidEmailFormat},
		{"no tld", "user@domain", "", d.ErrInvalidEmailFormat},
		{"too long", strings.Repeat("a", 251) + "@x.co", "", d.ErrEmailTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.NewEmail(tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, string(*got))
		})
	}
}

func TestNewWorkday(t *testing.T) {
	t.Run("single day", func(t *testing.T) {
		w, err := d.NewWorkday([]int{1}) // Monday
		require.NoError(t, err)
		assert.Equal(t, d.Monday, w)
	})

	t.Run("multiple days", func(t *testing.T) {
		w, err := d.NewWorkday([]int{1, 3, 5}) // Mon, Wed, Fri
		require.NoError(t, err)
		assert.Equal(t, d.Monday|d.Wednesday|d.Friday, w)
	})

	t.Run("all weekdays", func(t *testing.T) {
		w, err := d.NewWorkday([]int{1, 2, 3, 4, 5, 6, 7})
		require.NoError(t, err)
		assert.Equal(t, 7, len(w.ToSlice()))
	})

	t.Run("empty slice gives zero bitmask", func(t *testing.T) {
		w, err := d.NewWorkday([]int{})
		require.NoError(t, err)
		assert.Equal(t, d.Workday(0), w)
	})

	t.Run("day zero is invalid", func(t *testing.T) {
		_, err := d.NewWorkday([]int{0})
		assert.ErrorIs(t, err, d.ErrInvalidWorkday)
	})

	t.Run("day eight is invalid", func(t *testing.T) {
		_, err := d.NewWorkday([]int{8})
		assert.ErrorIs(t, err, d.ErrInvalidWorkday)
	})

	t.Run("negative day is invalid", func(t *testing.T) {
		_, err := d.NewWorkday([]int{-1})
		assert.ErrorIs(t, err, d.ErrInvalidWorkday)
	})
}

func TestWorkdayToSlice(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		input := []int{2, 4, 6}
		w, err := d.NewWorkday(input)
		require.NoError(t, err)
		assert.Equal(t, input, w.ToSlice())
	})

	t.Run("zero returns empty", func(t *testing.T) {
		assert.Empty(t, d.Workday(0).ToSlice())
	})

	t.Run("sunday is 7", func(t *testing.T) {
		w, err := d.NewWorkday([]int{7})
		require.NoError(t, err)
		assert.Equal(t, []int{7}, w.ToSlice())
	})
}

func TestNewWorkTime(t *testing.T) {
	valid := []string{"00:00", "09:00", "23:59", "9:05"}
	for _, v := range valid {
		t.Run("valid "+v, func(t *testing.T) {
			got, err := d.NewWorkTime(v)
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}

	invalid := []struct {
		name  string
		input string
		err   error
	}{
		{"empty", "", d.ErrEmptyWorkTime},
		{"only spaces", "  ", d.ErrEmptyWorkTime},
		{"bad format letters", "ab:cd", d.ErrInvalidWorkTimeFormat},
		{"out of range hours", "25:00", d.ErrInvalidWorkTimeFormat},
		{"out of range minutes", "12:60", d.ErrInvalidWorkTimeFormat},
		{"missing colon", "0900", d.ErrInvalidWorkTimeFormat},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.NewWorkTime(tt.input)
			assert.ErrorIs(t, err, tt.err)
			assert.Nil(t, got)
		})
	}
}
