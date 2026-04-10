package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"wishlist/internal/logic"
	"wishlist/internal/repo"
)

type PublicHandler struct {
	Store *repo.Store
}

func (h *PublicHandler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		Error(w, http.StatusBadRequest, "token is required")
		return
	}
	wl, err := h.Store.WishlistByPublicToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	items, err := h.Store.ItemsByWishlist(r.Context(), wl.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not load items")
		return
	}
	resp := logic.ToPublicResponse(wl, items)
	JSON(w, http.StatusOK, resp)
}

func (h *PublicHandler) ReserveItem(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		Error(w, http.StatusBadRequest, "token is required")
		return
	}
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil || itemID < 1 {
		Error(w, http.StatusBadRequest, "invalid item id")
		return
	}
	_, it, err := h.Store.ItemForPublicWishlist(r.Context(), token, itemID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist or item not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load item")
		return
	}
	if !logic.ItemReservable(it.ReservedAt) {
		Error(w, http.StatusConflict, "gift is already reserved")
		return
	}
	if err := h.Store.ReserveItemIfFree(r.Context(), it.WishlistID, itemID); err != nil {
		if errors.Is(err, repo.ErrAlreadyReserved) {
			Error(w, http.StatusConflict, "gift is already reserved")
			return
		}
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist or item not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not reserve gift")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
