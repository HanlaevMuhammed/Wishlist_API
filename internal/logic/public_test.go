package logic

import (
	"testing"
	"time"

	"wishlist/internal/models"
)

func TestItemReservable(t *testing.T) {
	if !ItemReservable(nil) {
		t.Fatal("nil reservedAt should be reservable")
	}
	now := time.Now()
	if ItemReservable(&now) {
		t.Fatal("non-nil reservedAt should not be reservable")
	}
}

func TestToPublicResponse(t *testing.T) {
	ev := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	w := models.Wishlist{Title: "NY", Description: "Party", EventDate: ev}
	now := time.Now()
	items := []models.WishlistItem{
		{ID: 1, Title: "Book", Description: "", ProductURL: "https://x", Priority: 5, ReservedAt: nil},
		{ID: 2, Title: "Mug", Description: "x", ProductURL: "", Priority: 3, ReservedAt: &now},
	}
	got := ToPublicResponse(w, items)
	if got.Title != "NY" || got.Description != "Party" || !got.EventDate.Equal(ev) {
		t.Fatalf("wishlist fields: %+v", got)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items len %d", len(got.Items))
	}
	if got.Items[0].ID != 1 || got.Items[0].Reserved {
		t.Fatalf("item 1: %+v", got.Items[0])
	}
	if got.Items[1].ID != 2 || !got.Items[1].Reserved {
		t.Fatalf("item 2: %+v", got.Items[1])
	}
}
