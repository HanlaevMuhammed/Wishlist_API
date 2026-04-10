package validation

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

const (
	MinPasswordLen = 8
	MaxTitleLen    = 500
	MaxDescLen     = 5000
	MaxURLLen      = 2048
)

func Email(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("email is required")
	}
	if strings.ContainsAny(s, "<>") {
		return fmt.Errorf("invalid email")
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return fmt.Errorf("invalid email")
	}
	return nil
}

func Password(s string) error {
	if len(s) < MinPasswordLen {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLen)
	}
	return nil
}

func Title(field, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(s) > MaxTitleLen {
		return fmt.Errorf("%s is too long", field)
	}
	return nil
}

func DescriptionOptional(s string) error {
	if len(s) > MaxDescLen {
		return fmt.Errorf("description is too long")
	}
	return nil
}

func ProductURLOptional(s string) error {
	if len(s) > MaxURLLen {
		return fmt.Errorf("product_url is too long")
	}
	return nil
}

func Priority(p int) error {
	if p < 1 || p > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}
	return nil
}

func EventDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("event_date is required")
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("event_date must be YYYY-MM-DD")
	}
	return t, nil
}
