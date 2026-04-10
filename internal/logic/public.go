package logic

import (
	"time"

	"wishlist/internal/models"
)

func ItemReservable(reservedAt *time.Time) bool {
	return reservedAt == nil
}

func ToPublicResponse(w models.Wishlist, items []models.WishlistItem) models.PublicWishlistResponse {
	out := models.PublicWishlistResponse{
		Title:       w.Title,
		Description: w.Description,
		EventDate:   w.EventDate,
		Items:       make([]models.PublicItemView, 0, len(items)),
	}
	for _, it := range items {
		reserved := !ItemReservable(it.ReservedAt)
		out.Items = append(out.Items, models.PublicItemView{
			ID:          it.ID,
			Title:       it.Title,
			Description: it.Description,
			ProductURL:  it.ProductURL,
			Priority:    it.Priority,
			Reserved:    reserved,
			ReservedAt:  it.ReservedAt,
		})
	}
	return out
}
