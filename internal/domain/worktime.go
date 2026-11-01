package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrEmptyWorkTime         = errors.New("work time is empty")
	ErrInvalidWorkTimeFormat = errors.New("work time must be in HH:MM format")
)

type WorkTime time.Time

var hoursAndMinutes = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

func NewWorkTime(raw string) (*WorkTime, error) {
	v := strings.TrimSpace(raw)

	if v == "" {
		return nil, ErrEmptyWorkTime
	}
	if !hoursAndMinutes.MatchString(v) {
		return nil, ErrInvalidWorkTimeFormat
	}
	t, err := time.Parse("15:04", v)
	if err != nil {
		return nil, ErrInvalidWorkTimeFormat
	}

	w := WorkTime(t)
	return &w, nil
}
