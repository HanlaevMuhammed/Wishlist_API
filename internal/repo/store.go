package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wishlist/internal/models"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	Pool *pgxpool.Pool
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2)
		RETURNING id, email, created_at`
	var u models.User
	err := s.Pool.QueryRow(ctx, q, email, passwordHash).Scan(&u.ID, &u.Email, &u.CreatedAt)
	return u, err
}

func (s *Store) UserByEmail(ctx context.Context, email string) (models.User, string, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	var u models.User
	var hash string
	err := s.Pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &hash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, "", ErrNotFound
	}
	return u, hash, err
}

func (s *Store) CreateWishlist(ctx context.Context, userID int64, title, description string, eventDate time.Time, publicToken string) (models.Wishlist, error) {
	const q = `INSERT INTO wishlists (user_id, title, description, event_date, public_token)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, description, event_date, public_token, created_at, updated_at`
	var w models.Wishlist
	err := s.Pool.QueryRow(ctx, q, userID, title, description, eventDate, publicToken).Scan(
		&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt,
	)
	return w, err
}

func (s *Store) WishlistsByUser(ctx context.Context, userID int64) ([]models.Wishlist, error) {
	const q = `SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists WHERE user_id = $1 ORDER BY event_date ASC, id ASC`
	rows, err := s.Pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Wishlist
	for rows.Next() {
		var w models.Wishlist
		if err := rows.Scan(&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) WishlistByIDForUser(ctx context.Context, id, userID int64) (models.Wishlist, error) {
	const q = `SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists WHERE id = $1 AND user_id = $2`
	var w models.Wishlist
	err := s.Pool.QueryRow(ctx, q, id, userID).Scan(
		&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Wishlist{}, ErrNotFound
	}
	return w, err
}

func (s *Store) UpdateWishlist(ctx context.Context, id, userID int64, title, description *string, eventDate *time.Time) (models.Wishlist, error) {
	w, err := s.WishlistByIDForUser(ctx, id, userID)
	if err != nil {
		return models.Wishlist{}, err
	}
	if title != nil {
		w.Title = *title
	}
	if description != nil {
		w.Description = *description
	}
	if eventDate != nil {
		w.EventDate = *eventDate
	}
	const q = `UPDATE wishlists SET title = $1, description = $2, event_date = $3, updated_at = NOW()
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, title, description, event_date, public_token, created_at, updated_at`
	err = s.Pool.QueryRow(ctx, q, w.Title, w.Description, w.EventDate, id, userID).Scan(
		&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt,
	)
	return w, err
}

func (s *Store) DeleteWishlist(ctx context.Context, id, userID int64) error {
	cmd, err := s.Pool.Exec(ctx, `DELETE FROM wishlists WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) WishlistByPublicToken(ctx context.Context, token string) (models.Wishlist, error) {
	const q = `SELECT id, user_id, title, description, event_date, public_token, created_at, updated_at
		FROM wishlists WHERE public_token = $1`
	var w models.Wishlist
	err := s.Pool.QueryRow(ctx, q, token).Scan(
		&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Wishlist{}, ErrNotFound
	}
	return w, err
}

func (s *Store) ItemsByWishlist(ctx context.Context, wishlistID int64) ([]models.WishlistItem, error) {
	const q = `SELECT id, wishlist_id, title, description, product_url, priority, reserved_at, created_at, updated_at
		FROM wishlist_items WHERE wishlist_id = $1 ORDER BY priority DESC, id ASC`
	rows, err := s.Pool.Query(ctx, q, wishlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WishlistItem
	for rows.Next() {
		var it models.WishlistItem
		if err := rows.Scan(&it.ID, &it.WishlistID, &it.Title, &it.Description, &it.ProductURL, &it.Priority, &it.ReservedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) CreateItem(ctx context.Context, wishlistID int64, title, description, productURL string, priority int) (models.WishlistItem, error) {
	const q = `INSERT INTO wishlist_items (wishlist_id, title, description, product_url, priority)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, wishlist_id, title, description, product_url, priority, reserved_at, created_at, updated_at`
	var it models.WishlistItem
	err := s.Pool.QueryRow(ctx, q, wishlistID, title, description, productURL, priority).Scan(
		&it.ID, &it.WishlistID, &it.Title, &it.Description, &it.ProductURL, &it.Priority, &it.ReservedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	return it, err
}

func (s *Store) ItemByIDForWishlist(ctx context.Context, itemID, wishlistID int64) (models.WishlistItem, error) {
	const q = `SELECT id, wishlist_id, title, description, product_url, priority, reserved_at, created_at, updated_at
		FROM wishlist_items WHERE id = $1 AND wishlist_id = $2`
	var it models.WishlistItem
	err := s.Pool.QueryRow(ctx, q, itemID, wishlistID).Scan(
		&it.ID, &it.WishlistID, &it.Title, &it.Description, &it.ProductURL, &it.Priority, &it.ReservedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.WishlistItem{}, ErrNotFound
	}
	return it, err
}

func (s *Store) UpdateItem(ctx context.Context, itemID, wishlistID int64, title, description, productURL *string, priority *int) (models.WishlistItem, error) {
	it, err := s.ItemByIDForWishlist(ctx, itemID, wishlistID)
	if err != nil {
		return models.WishlistItem{}, err
	}
	if title != nil {
		it.Title = *title
	}
	if description != nil {
		it.Description = *description
	}
	if productURL != nil {
		it.ProductURL = *productURL
	}
	if priority != nil {
		it.Priority = *priority
	}
	const q = `UPDATE wishlist_items SET title = $1, description = $2, product_url = $3, priority = $4, updated_at = NOW()
		WHERE id = $5 AND wishlist_id = $6
		RETURNING id, wishlist_id, title, description, product_url, priority, reserved_at, created_at, updated_at`
	err = s.Pool.QueryRow(ctx, q, it.Title, it.Description, it.ProductURL, it.Priority, itemID, wishlistID).Scan(
		&it.ID, &it.WishlistID, &it.Title, &it.Description, &it.ProductURL, &it.Priority, &it.ReservedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	return it, err
}

func (s *Store) DeleteItem(ctx context.Context, itemID, wishlistID int64) error {
	cmd, err := s.Pool.Exec(ctx, `DELETE FROM wishlist_items WHERE id = $1 AND wishlist_id = $2`, itemID, wishlistID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ItemForPublicWishlist(ctx context.Context, publicToken string, itemID int64) (models.Wishlist, models.WishlistItem, error) {
	const q = `
		SELECT w.id, w.user_id, w.title, w.description, w.event_date, w.public_token, w.created_at, w.updated_at,
			i.id, i.wishlist_id, i.title, i.description, i.product_url, i.priority, i.reserved_at, i.created_at, i.updated_at
		FROM wishlists w
		JOIN wishlist_items i ON i.wishlist_id = w.id
		WHERE w.public_token = $1 AND i.id = $2`
	var w models.Wishlist
	var it models.WishlistItem
	err := s.Pool.QueryRow(ctx, q, publicToken, itemID).Scan(
		&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.PublicToken, &w.CreatedAt, &w.UpdatedAt,
		&it.ID, &it.WishlistID, &it.Title, &it.Description, &it.ProductURL, &it.Priority, &it.ReservedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Wishlist{}, models.WishlistItem{}, ErrNotFound
	}
	return w, it, err
}

func (s *Store) ReserveItemIfFree(ctx context.Context, wishlistID, itemID int64) error {
	cmd, err := s.Pool.Exec(ctx,
		`UPDATE wishlist_items SET reserved_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND wishlist_id = $2 AND reserved_at IS NULL`,
		itemID, wishlistID,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		it, e := s.ItemByIDForWishlist(ctx, itemID, wishlistID)
		if e != nil {
			return ErrNotFound
		}
		if it.ReservedAt != nil {
			return ErrAlreadyReserved
		}
		return ErrNotFound
	}
	return nil
}

var ErrAlreadyReserved = errors.New("already reserved")
