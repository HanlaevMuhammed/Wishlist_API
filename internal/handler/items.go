package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"wishlist/internal/models"
	"wishlist/internal/repo"
	"wishlist/internal/validation"
)

type ItemHandler struct {
	Store *repo.Store
}

type createItemReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ProductURL  string `json:"product_url"`
	Priority    int    `json:"priority"`
}

type patchItemReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ProductURL  *string `json:"product_url"`
	Priority    *int    `json:"priority"`
}

func (h *ItemHandler) wishlistID(r *http.Request) (int64, int64, bool) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		return 0, 0, false
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "wishlistID"), 10, 64)
	if err != nil || id < 1 {
		return uid, 0, false
	}
	return uid, id, true
}

func (h *ItemHandler) ensureOwner(ctx context.Context, uid, wlID int64) (models.Wishlist, error) {
	wl, err := h.Store.WishlistByIDForUser(ctx, wlID, uid)
	if err != nil {
		return models.Wishlist{}, err
	}
	return wl, nil
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, wlID, ok := h.wishlistID(r)
	if !ok {
		if _, hasUser := UserIDFromContext(r.Context()); !hasUser {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	if _, err := h.ensureOwner(r.Context(), uid, wlID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	var req createItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validation.Title("title", req.Title); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if err := validation.DescriptionOptional(req.Description); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.ProductURLOptional(req.ProductURL); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Priority == 0 {
		req.Priority = 1
	}
	if err := validation.Priority(req.Priority); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	it, err := h.Store.CreateItem(r.Context(), wlID, req.Title, req.Description, req.ProductURL, req.Priority)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not create item")
		return
	}
	JSON(w, http.StatusCreated, it)
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, wlID, ok := h.wishlistID(r)
	if !ok {
		if _, hasUser := UserIDFromContext(r.Context()); !hasUser {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	if _, err := h.ensureOwner(r.Context(), uid, wlID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	items, err := h.Store.ItemsByWishlist(r.Context(), wlID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not load items")
		return
	}
	if items == nil {
		items = []models.WishlistItem{}
	}
	JSON(w, http.StatusOK, items)
}

func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, wlID, ok := h.wishlistID(r)
	if !ok {
		if _, hasUser := UserIDFromContext(r.Context()); !hasUser {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil || itemID < 1 {
		Error(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if _, err := h.ensureOwner(r.Context(), uid, wlID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	it, err := h.Store.ItemByIDForWishlist(r.Context(), itemID, wlID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "item not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load item")
		return
	}
	JSON(w, http.StatusOK, it)
}

func (h *ItemHandler) Patch(w http.ResponseWriter, r *http.Request) {
	uid, wlID, ok := h.wishlistID(r)
	if !ok {
		if _, hasUser := UserIDFromContext(r.Context()); !hasUser {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil || itemID < 1 {
		Error(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if _, err := h.ensureOwner(r.Context(), uid, wlID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	var req patchItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Title == nil && req.Description == nil && req.ProductURL == nil && req.Priority == nil {
		Error(w, http.StatusBadRequest, "no fields to update")
		return
	}
	if req.Title != nil {
		if err := validation.Title("title", *req.Title); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		trimmed := strings.TrimSpace(*req.Title)
		req.Title = &trimmed
	}
	if req.Description != nil {
		if err := validation.DescriptionOptional(*req.Description); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.ProductURL != nil {
		if err := validation.ProductURLOptional(*req.ProductURL); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.Priority != nil {
		if err := validation.Priority(*req.Priority); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	it, err := h.Store.UpdateItem(r.Context(), itemID, wlID, req.Title, req.Description, req.ProductURL, req.Priority)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "item not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not update item")
		return
	}
	JSON(w, http.StatusOK, it)
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, wlID, ok := h.wishlistID(r)
	if !ok {
		if _, hasUser := UserIDFromContext(r.Context()); !hasUser {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil || itemID < 1 {
		Error(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if _, err := h.ensureOwner(r.Context(), uid, wlID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	if err := h.Store.DeleteItem(r.Context(), itemID, wlID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "item not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not delete item")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
