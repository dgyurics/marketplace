package routes

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type UserRoutes struct {
	router
	userService    services.UserService
	jwtService     services.JWTService
	refreshService services.RefreshService
	config         types.AuthConfig
}

func NewUserRoutes(
	userService services.UserService,
	jwtService services.JWTService,
	refreshService services.RefreshService,
	config types.AuthConfig,
	router router) *UserRoutes {
	return &UserRoutes{
		router:         router,
		userService:    userService,
		jwtService:     jwtService,
		refreshService: refreshService,
		config:         config,
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	if email == "" || strings.Contains(email, " ") || len(email) > 254 {
		return false
	}

	// Use a simple regex for a quick check
	if !emailRegex.MatchString(email) {
		return false
	}

	// Use Go's built-in parser for more strict validation
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (h *UserRoutes) Login(w http.ResponseWriter, r *http.Request) {
	var credentials types.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if credentials.Email == "" || !isValidEmail(credentials.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email required")
		return
	}

	if credentials.Password == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Password required")
		return
	}

	// Verify user credentials
	usr, err := h.userService.Login(r.Context(), &credentials)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate access token
	accessToken, err := h.jwtService.GenerateToken(*usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate refresh token
	refreshToken, err := h.refreshService.GenerateToken()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store refresh token
	if err := h.refreshService.StoreToken(r.Context(), usr.ID, refreshToken); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token":          accessToken,
		"refresh_token":  refreshToken,
		"requires_setup": usr.RequiresSetup,
	})
}

// RefreshToken generates a new access token using a valid refresh token
func (h *UserRoutes) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
		// UserID       string `json:"user_id"` // TODO
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}
	if requestBody.RefreshToken == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Refresh token required")
		return
	}

	// Parse the refresh token
	// FIXME better error handling
	// does not distinguish between bad input and system errors
	usr, err := h.refreshService.VerifyToken(r.Context(), requestBody.RefreshToken)
	if err != nil {
		u.RespondWithError(w, r, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// TODO cycle refresh tokens
	// revoke the just validated refresh token and generate a new one

	// Generate a new access token
	accessToken, err := h.jwtService.GenerateToken(*usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, map[string]string{
		"token":         accessToken,
		"refresh_token": requestBody.RefreshToken,
	})
}

func (h *UserRoutes) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.refreshService.RevokeTokens(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *UserRoutes) CreateGuestUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create guest user
	usr := types.User{}
	err := h.userService.CreateGuest(ctx, &usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate access token
	accessToken, err := h.jwtService.GenerateToken(usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate refresh token
	var token string
	token, err = h.refreshService.GenerateToken()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store refresh token
	err = h.refreshService.StoreToken(r.Context(), usr.ID, token)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"token":         accessToken,
		"refresh_token": token,
	})
}

// UpdatedCredentials is used used for overriding the default admin account, on system setup
// FIXME allows for any email to be used (without verification)
func (h *UserRoutes) UpdatedCredentials(w http.ResponseWriter, r *http.Request) {
	// verify request contains email and password
	var credentials types.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if credentials.Email == "" || !isValidEmail(credentials.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email is required")
		return
	}
	if credentials.Password == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Password is required")
		return
	}

	// Update user email and password
	usr, err := h.userService.SetCredentials(r.Context(), credentials)
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, "User with this email already exists")
		return
	}
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate new access token
	accessToken, err := h.jwtService.GenerateToken(usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate new refresh token
	var token string
	token, err = h.refreshService.GenerateToken()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store refresh token
	err = h.refreshService.StoreToken(r.Context(), usr.ID, token)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"token":         accessToken,
		"refresh_token": token,
	})
}

func (h *UserRoutes) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	params := u.ParsePaginationParams(r, 1, 100)
	users, err := h.userService.GetAllUsers(r.Context(), params.Page, params.Limit)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, users)
}

func (h *UserRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/users/login", h.Login).Methods(http.MethodPost)                // TODO rate limit to prevent abuse
	h.muxRouter.HandleFunc("/users/refresh-token", h.RefreshToken).Methods(http.MethodPost) // TODO rate limit to prevent abuse
	h.muxRouter.HandleFunc("/users/guest", h.CreateGuestUser).Methods(http.MethodPost)      // TODO rate limit to prevent abuse
	h.muxRouter.Handle("/users/credentials", h.secureAdmin(h.UpdatedCredentials)).Methods(http.MethodPut)
	h.muxRouter.Handle("/users/logout", h.secure(h.Logout)).Methods(http.MethodPost)
	h.muxRouter.Handle("/users", h.secureAdmin(h.GetAllUsers)).Methods(http.MethodGet)
}
