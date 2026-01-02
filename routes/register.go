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

type RegisterRoutes struct {
	router
	userService     services.UserService
	jwtService      services.JWTService
	refreshService  services.RefreshService
	emailService    services.EmailService
	templateService services.TemplateService
	baseURL         string // TODO move BaseURL INSIDE tempalteService
}

func NewRegisterRoutes(userService services.UserService,
	jwtService services.JWTService,
	refreshService services.RefreshService,
	emailService services.EmailService,
	templateService services.TemplateService,
	baseURL string,
	router router) *RegisterRoutes {
	return &RegisterRoutes{
		router:          router,
		userService:     userService,
		jwtService:      jwtService,
		refreshService:  refreshService,
		emailService:    emailService,
		templateService: templateService,
		baseURL:         baseURL,
	}
}

func (h *RegisterRoutes) Register(w http.ResponseWriter, r *http.Request) {
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
		Email:    strings.ToLower(reqBody.Email),
		Password: reqBody.Password,
		Role:     types.RoleUser,
		Verified: false,
	}
	err := h.userService.CreateUser(r.Context(), &usr)
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, "email already in-use")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// create registration code
	code, err := h.userService.CreateRegistrationCode(r.Context(), usr.ID)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// email customer email verification link
	go func(email, code string) {
		detailsLink := fmt.Sprintf("%s/auth?registration-code=%s",
			h.baseURL,
			url.QueryEscape(code))
		data := map[string]string{
			"DetailsLink": detailsLink,
		}
		body, err := h.templateService.RenderToString(services.EmailVerification, data)
		if err != nil {
			slog.Error("Error loading email template: ", "error", err)
			return
		}
		payload := &types.Email{
			To:      []string{email},
			Subject: "Email Verification",
			Body:    body,
			IsHTML:  true,
		}
		if err := h.emailService.Send(payload); err != nil {
			slog.Error("Error sending new user registration email: ", "email", email, "error", err)
		}
	}(usr.Email, code)

	u.RespondSuccess(w)
}

// RegisterConfirm handles the confirmation of a user's registration code
// It marks the user as verified if the registration code is valid. (Afterwards they can log in)
// It returns a 400 status code if the registration code is invalid or expired
func (h *RegisterRoutes) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	// extract registration_code
	var reqBody struct {
		RegistrationCode string `json:"registration_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	// confirm the registration code (mark user as verified if valid)
	err := h.userService.ConfirmRegistrationCode(r.Context(), reqBody.RegistrationCode)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *RegisterRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/register", h.limit(h.Register, 2, time.Hour*6)).Methods(http.MethodPost)
	h.muxRouter.Handle("/register/confirm", h.limit(h.RegisterConfirm, 2, time.Hour*6)).Methods(http.MethodPost)
}
