package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type UserRoutes struct {
	router
	userService services.UserService
	authService services.AuthService
}

func NewUserRoutes(userService services.UserService, authService services.AuthService, router router) *UserRoutes {
	return &UserRoutes{
		router:      router,
		userService: userService,
		authService: authService,
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

func (h *UserRoutes) Register(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	credentials.Email = strings.ToLower(credentials.Email) // store email in lowercase

	if credentials.Email == "" || !isValidEmail(credentials.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email is required")
		return
	}

	if credentials.Password == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Password is required")
		return
	}

	// Check if registration code is required and if the invite code is valid
	inviteCodeReq := u.IsFeatureEnabled("REQUIRE_INVITE_CODE")
	if inviteCodeReq && len(credentials.InviteCode) != 6 {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invite code is required")
		return
	}

	// fetch the invite code
	valid, err := h.authService.ValidateInviteCode(r.Context(), credentials.InviteCode, inviteCodeReq)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if !valid {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid invite code")
		return
	}

	usr := models.User{
		Email:    credentials.Email,
		Password: credentials.Password,
	}

	// Create the user
	if err := h.userService.CreateUser(r.Context(), &usr); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Mark the invite code as used
	if err := h.authService.StoreInviteCode(r.Context(), credentials.InviteCode, true); err != nil {
		slog.Error("Failed to update invite code", "code", credentials.InviteCode, "error", err.Error())
	}

	// Generate access token
	accessToken, err := h.authService.GenerateAccessToken(usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate refresh token
	var token string
	token, err = h.authService.GenerateRefreshToken()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store refresh token
	err = h.authService.StoreRefreshToken(r.Context(), usr.ID, token)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"token":         accessToken,
		"refresh_token": token,
	})
}

func (h *UserRoutes) Login(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
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
	if err != nil {
		u.RespondWithError(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate access token
	accessToken, err := h.authService.GenerateAccessToken(*usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate refresh token
	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store refresh token
	if err := h.authService.StoreRefreshToken(r.Context(), usr.ID, refreshToken); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, map[string]string{
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// Exists checks if a user with the given email exists
func (h *UserRoutes) Exists(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	// Validate the email
	credentials.Email = strings.ToLower(credentials.Email)
	if credentials.Email == "" || !isValidEmail(credentials.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email is required")
		return
	}

	// Check if the user exists
	usr, err := h.userService.GetUserByEmail(r.Context(), credentials.Email)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	usrExists := usr != nil
	u.RespondWithJSON(w, http.StatusOK, map[string]bool{"exists": usrExists})
}

// RefreshToken generates a new access token using a valid refresh token
func (h *UserRoutes) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}
	if requestBody.RefreshToken == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Refresh token required")
		return
	}

	// Validate the refresh token
	user, err := h.authService.ValidateRefreshToken(r.Context(), requestBody.RefreshToken)
	if err != nil {
		u.RespondWithError(w, r, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// TODO cycle refresh tokens
	// revoke the just validated refresh token and generate a new one

	// Generate a new access token
	accessToken, err := h.authService.GenerateAccessToken(user)
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
	if err := h.authService.RevokeRefreshTokens(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *UserRoutes) GetAddresses(w http.ResponseWriter, r *http.Request) {
	addresses, err := h.userService.GetAddresses(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, addresses)
}

func (h *UserRoutes) CreateAddress(w http.ResponseWriter, r *http.Request) {
	var address models.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := h.userService.CreateAddress(r.Context(), &address); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, address)
}

func (h *UserRoutes) RemoveAddress(w http.ResponseWriter, r *http.Request) {
	addressID := mux.Vars(r)["id"]
	if err := h.userService.RemoveAddress(r.Context(), addressID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
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

func (h *UserRoutes) GenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	// Generate a new invite code
	code, err := h.authService.GenerateInviteCode(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	// Store the invite code in the database
	if err = h.authService.StoreInviteCode(r.Context(), code, false); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, code)
}

func (h *UserRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/users/register", h.Register).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/login", h.Login).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/logout", h.Logout).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/refresh-token", h.RefreshToken).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/exists", h.Exists).Methods(http.MethodPost)
	h.muxRouter.Handle("/users/addresses", h.secure(h.GetAddresses)).Methods(http.MethodGet)
	h.muxRouter.Handle("/users/addresses", h.secure(h.CreateAddress)).Methods(http.MethodPost)
	h.muxRouter.Handle("/users/addresses/{id}", h.secure(h.RemoveAddress)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/users", h.secureAdmin(h.GetAllUsers)).Methods(http.MethodGet)
	h.muxRouter.Handle("/users/invite", h.secureAdmin(h.GenerateInviteCode)).Methods(http.MethodPost)
	// router.HandleFunc("/users/profile", GetProfile).Methods("GET")
	// router.HandleFunc("/users/update-profile", UpdateProfile).Methods("POST")
	// router.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE")
}
