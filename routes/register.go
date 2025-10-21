package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type RegisterRoutes struct {
	router
	regService     services.RegisterService
	userService    services.UserService
	jwtService     services.JWTService
	refreshService services.RefreshService
}

func NewRegisterRoutes(regService services.RegisterService,
	userService services.UserService,
	jwtService services.JWTService,
	refreshService services.RefreshService,
	router router) *RegisterRoutes {
	return &RegisterRoutes{
		router:         router,
		regService:     regService,
		userService:    userService,
		jwtService:     jwtService,
		refreshService: refreshService,
	}
}

func (h *RegisterRoutes) Register(w http.ResponseWriter, r *http.Request) {
	var reqBody types.Credential
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}
	if reqBody.Email == "" || !isValidEmail(reqBody.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email is required")
		return
	}

	// create entry in pending_users table
	regCode, err := h.regService.Register(r.Context(), strings.ToLower(reqBody.Email))
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, "User with this email already exists")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO send email using regCode
	slog.DebugContext(r.Context(), "Sending registration email", "email", reqBody.Email, "code", regCode)

	u.RespondSuccess(w)
}

func (h *RegisterRoutes) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	var reqBody types.Credential
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	reqBody.Email = strings.ToLower(reqBody.Email) // store email in lowercase

	// Basic validation
	if reqBody.Email == "" || !isValidEmail(reqBody.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email is required")
		return
	}
	if reqBody.RegistrationCode == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Registration code is required")
		return
	}
	if reqBody.Password == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Password is required")
		return
	}

	// Confirm email and code has matching entry in pending_users table
	err := h.regService.RegisterConfirm(r.Context(), reqBody.Email, reqBody.RegistrationCode)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid email or registration code")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Create the user
	usr := types.User{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}
	err = h.userService.CreateUser(r.Context(), &usr)
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, "User with this email already exists")
		return
	}
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

	u.RespondSuccess(w)
}

func (h *RegisterRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/register", h.limit(h.Register, 2, time.Hour)).Methods(http.MethodPost)
	h.muxRouter.Handle("/register/confirm", h.limit(h.RegisterConfirm, 2, time.Hour)).Methods(http.MethodPost)
}
