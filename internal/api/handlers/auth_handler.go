package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jsi/ibs-doc-engine/internal/api"
)

// AuthConfig holds credentials for the built-in admin login.
type AuthConfig struct {
	AdminUsername string
	AdminPassword string
	AdminAPIKey   string // plaintext API key returned as token on successful login
}

// AuthHandler handles login/logout.
type AuthHandler struct {
	cfg AuthConfig
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(cfg AuthConfig) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

// Routes returns a chi.Router with auth routes.
func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	return r
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// Login handles POST /login — validates credentials and returns an API key as token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if body.Username == "" || body.Password == "" {
		api.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "username and password are required")
		return
	}

	if body.Username != h.cfg.AdminUsername || body.Password != h.cfg.AdminPassword {
		api.JSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
		return
	}

	api.JSON(w, http.StatusOK, loginResponse{
		Token:    h.cfg.AdminAPIKey,
		Username: body.Username,
	}, nil)
}
