package routes

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type RegistrationRoutes struct {
	router
	userService         services.UserService
	registrationService services.RegistrationService
	jwtService          services.JWTService
	refreshService      services.RefreshService
	notificationService services.NotificationService
}

func NewRegisterRoutes(
	userService services.UserService,
	registrationService services.RegistrationService,
	jwtService services.JWTService,
	refreshService services.RefreshService,
	notificationService services.NotificationService,
	router router) *RegistrationRoutes {
	return &RegistrationRoutes{
		router:              router,
		userService:         userService,
		registrationService: registrationService,
		jwtService:          jwtService,
		refreshService:      refreshService,
		notificationService: notificationService,
	}
}

func (h *RegistrationRoutes) Register(w http.ResponseWriter, r *http.Request) {
	var reqBody types.Credential
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}
	if reqBody.Email == "" || !isValidEmail(reqBody.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "email is required")
		return
	}
	if reqBody.Password == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "password is required")
		return
	}

	// create new user
	usr := types.User{
		Email:    u.StringPtr(strings.ToLower(reqBody.Email)),
		Password: &reqBody.Password,
		Role:     types.RoleUser,
		Verified: false,
	}
	err := h.userService.CreateUser(r.Context(), &usr)
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// create registration code
	code, err := h.registrationService.CreateCode(r.Context(), usr.ID, time.Now().UTC().Add(24*time.Hour))
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// email customer email verification link
	go func(email, code string) {
		detailsLink := fmt.Sprintf("%s?registration-code=%s", h.notificationService.BaseURL(), url.QueryEscape(code))
		data := map[string]string{
			"DetailsLink": detailsLink,
		}
		if err := h.notificationService.SendEmail(email, "Email Verification", services.EmailVerification, data); err != nil {
			slog.Error("Error sending new user registration email: ", "email", email, "error", err)
		}
	}(*usr.Email, code)

	u.RespondSuccess(w)
}

func (h *RegistrationRoutes) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RegistrationCode string `json:"registration_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	// verify registration code
	usr, err := h.registrationService.VerifyCode(r.Context(), reqBody.RegistrationCode)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate new access token
	accessToken, err := h.jwtService.GenerateToken(*usr)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate new refresh refreshToken
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

	u.RespondWithJSON(w, http.StatusCreated, types.TokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *RegistrationRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/register", h.limit(h.Register, 2, time.Hour*6)).Methods(http.MethodPost)
	h.muxRouter.Handle("/register/confirm", h.limit(h.RegisterConfirm, 2, time.Hour*6)).Methods(http.MethodPost)
}
