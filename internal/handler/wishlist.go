package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"wishlist/internal/auth"
	"wishlist/internal/models"
	"wishlist/internal/repo"
	"wishlist/internal/validation"
)

type WishlistHandler struct {
	Store *repo.Store
}

type createWishlistReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
}

type patchWishlistReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	EventDate   *string `json:"event_date"`
}

func (h *WishlistHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createWishlistReq
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
	ev, err := validation.EventDate(req.EventDate)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	token, err := auth.PublicToken()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not generate public link")
		return
	}
	wl, err := h.Store.CreateWishlist(r.Context(), uid, req.Title, req.Description, ev, token)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not create wishlist")
		return
	}
	JSON(w, http.StatusCreated, wl)
}

func (h *WishlistHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.Store.WishlistsByUser(r.Context(), uid)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not load wishlists")
		return
	}
	if list == nil {
		list = []models.Wishlist{}
	}
	JSON(w, http.StatusOK, list)
}

func (h *WishlistHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	wl, err := h.Store.WishlistByIDForUser(r.Context(), id, uid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load wishlist")
		return
	}
	JSON(w, http.StatusOK, wl)
}

func (h *WishlistHandler) Patch(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	var req patchWishlistReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Title == nil && req.Description == nil && req.EventDate == nil {
		Error(w, http.StatusBadRequest, "no fields to update")
		return
	}
	var titlePtr, descPtr *string
	var eventPtr *time.Time
	if req.Title != nil {
		if err := validation.Title("title", *req.Title); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		trimmed := strings.TrimSpace(*req.Title)
		req.Title = &trimmed
		titlePtr = req.Title
	}
	if req.Description != nil {
		if err := validation.DescriptionOptional(*req.Description); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		descPtr = req.Description
	}
	if req.EventDate != nil {
		ev, err := validation.EventDate(*req.EventDate)
		if err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		eventPtr = &ev
	}
	wl, err := h.Store.UpdateWishlist(r.Context(), id, uid, titlePtr, descPtr, eventPtr)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not update wishlist")
		return
	}
	JSON(w, http.StatusOK, wl)
}

func (h *WishlistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		Error(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	if err := h.Store.DeleteWishlist(r.Context(), id, uid); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusNotFound, "wishlist not found")
			return
		}
		Error(w, http.StatusInternalServerError, "could not delete wishlist")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
