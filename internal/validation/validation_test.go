package validation

import "testing"

func TestEmail(t *testing.T) {
	if err := Email(""); err == nil {
		t.Fatal("empty email")
	}
	if err := Email("not-an-email"); err == nil {
		t.Fatal("invalid email")
	}
	if err := Email("user@example.com"); err != nil {
		t.Fatal(err)
	}
}

func TestPassword(t *testing.T) {
	if err := Password("short"); err == nil {
		t.Fatal("short password")
	}
	if err := Password("longenough"); err != nil {
		t.Fatal(err)
	}
}

func TestPriority(t *testing.T) {
	if err := Priority(0); err == nil {
		t.Fatal("0")
	}
	if err := Priority(11); err == nil {
		t.Fatal("11")
	}
	if err := Priority(5); err != nil {
		t.Fatal(err)
	}
}

func TestEventDate(t *testing.T) {
	_, err := EventDate("2026-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EventDate("15-01-2026"); err == nil {
		t.Fatal("wrong format accepted")
	}
}
