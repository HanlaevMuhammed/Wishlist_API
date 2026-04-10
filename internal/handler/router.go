package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"wishlist/internal/config"
	"wishlist/internal/repo"
)

func NewRouter(cfg config.Config, store *repo.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	ah := &AuthHandler{Store: store, Config: cfg}
	wh := &WishlistHandler{Store: store}
	ih := &ItemHandler{Store: store}
	ph := &PublicHandler{Store: store}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", ah.Register)
		r.Post("/auth/login", ah.Login)

		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(cfg.JWTSecret))
			r.Post("/wishlists", wh.Create)
			r.Get("/wishlists", wh.List)
			r.Get("/wishlists/{id}", wh.Get)
			r.Patch("/wishlists/{id}", wh.Patch)
			r.Delete("/wishlists/{id}", wh.Delete)

			r.Post("/wishlists/{wishlistID}/items", ih.Create)
			r.Get("/wishlists/{wishlistID}/items", ih.List)
			r.Get("/wishlists/{wishlistID}/items/{itemID}", ih.Get)
			r.Patch("/wishlists/{wishlistID}/items/{itemID}", ih.Patch)
			r.Delete("/wishlists/{wishlistID}/items/{itemID}", ih.Delete)
		})
	})

	r.Get("/public/v1/wishlists/{token}", ph.GetWishlist)
	r.Post("/public/v1/wishlists/{token}/items/{itemID}/reserve", ph.ReserveItem)

	return r
}
