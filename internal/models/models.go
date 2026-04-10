package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Wishlist struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	PublicToken string    `json:"public_token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WishlistItem struct {
	ID          int64      `json:"id"`
	WishlistID  int64      `json:"wishlist_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ProductURL  string     `json:"product_url"`
	Priority    int        `json:"priority"`
	ReservedAt  *time.Time `json:"reserved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PublicWishlistResponse struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	EventDate   time.Time        `json:"event_date"`
	Items       []PublicItemView `json:"items"`
}

type PublicItemView struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ProductURL  string     `json:"product_url"`
	Priority    int        `json:"priority"`
	Reserved    bool       `json:"reserved"`
	ReservedAt  *time.Time `json:"reserved_at,omitempty"`
}
