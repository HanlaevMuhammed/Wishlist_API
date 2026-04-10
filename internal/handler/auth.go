package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"wishlist/internal/auth"
	"wishlist/internal/config"
	"wishlist/internal/repo"
	"wishlist/internal/validation"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Store  *repo.Store
	Config config.Config
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authUserResp struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type authResp struct {
	User  authUserResp `json:"user"`
	Token string       `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validation.Email(req.Email); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if err := validation.Password(req.Password); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not process password")
		return
	}
	u, err := h.Store.CreateUser(r.Context(), req.Email, string(hash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			Error(w, http.StatusConflict, "email already registered")
			return
		}
		Error(w, http.StatusInternalServerError, "could not create user")
		return
	}
	token, err := auth.SignToken(u.ID, h.Config.JWTSecret, h.Config.JWTExpiration)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	JSON(w, http.StatusCreated, authResp{
		User: authUserResp{
			ID:    u.ID,
			Email: u.Email,
		},
		Token: token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validation.Email(req.Email); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Password == "" {
		Error(w, http.StatusBadRequest, "password is required")
		return
	}
	u, hash, err := h.Store.UserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		Error(w, http.StatusInternalServerError, "could not load user")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	token, err := auth.SignToken(u.ID, h.Config.JWTSecret, h.Config.JWTExpiration)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	JSON(w, http.StatusOK, authResp{
		User: authUserResp{
			ID:    u.ID,
			Email: u.Email,
		},
		Token: token,
	})
}
