package routes

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type PasswordRoutes struct {
	router
	servicePassword services.PasswordService
	serviceUser     services.UserService
	serviceEmail    services.EmailService
	serviceTemplate services.TemplateService
	baseUrl         string
}

func NewPasswordRoutes(
	servicePR services.PasswordService,
	serviceUsr services.UserService,
	serviceEmail services.EmailService,
	serviceTmp services.TemplateService,
	baseUrl string,
	router router,
) *PasswordRoutes {
	return &PasswordRoutes{
		router:          router,
		servicePassword: servicePR,
		serviceUser:     serviceUsr,
		serviceEmail:    serviceEmail,
		serviceTemplate: serviceTmp,
		baseUrl:         baseUrl,
	}
}

func (h *PasswordRoutes) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var credentials types.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if credentials.Email == "" || !isValidEmail(credentials.Email) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Email is required")
		return
	}

	// Check if the user exists
	usr, err := h.serviceUser.GetUserByEmail(r.Context(), credentials.Email)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// If user does not exist, respond with success to avoid email enumeration
	if usr == nil {
		slog.Warn("Password reset requested for non-existing email: ", "email", credentials.Email)
		u.RespondSuccess(w)
		return
	}

	// Generate password reset code
	code, err := h.servicePassword.GenerateResetCode(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store password reset code
	if err := h.servicePassword.StoreResetCode(r.Context(), code, usr.ID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Send password reset email
	go func(recEmail, code string) {
		data := map[string]string{
			"ResetLink": fmt.Sprintf("%s/auth/email/%s/password-reset/%s", h.baseUrl, recEmail, code),
		}
		body, err := h.serviceTemplate.RenderToString(services.PasswordReset, data)
		if err != nil {
			slog.Error("Error loading email template: ", "error", err)
			return
		}
		email := &types.Email{
			To:      []string{recEmail},
			Subject: "Password Reset",
			Body:    body,
			IsHTML:  true,
		}
		if err := h.serviceEmail.Send(email); err != nil {
			slog.Error("Error sending password reset email: ", "error", err)
		}
	}(usr.Email, code)

	u.RespondSuccess(w)
}

// TODO return unique status code when code has expired (allows frontend to show specific message)
// TODO redesign to work with just the code (no email) - code should be unique enough
func (h *PasswordRoutes) ResetPasswordConfirm(w http.ResponseWriter, r *http.Request) {
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

	if credentials.ResetCode == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Reset code is required")
		return
	}

	// Validate email and code match, exists, and is not expired
	valid, err := h.servicePassword.ValidateResetCode(r.Context(), credentials.ResetCode, credentials.Email)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if !valid {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid reset code")
		return
	}

	// Reset the password
	if err := h.servicePassword.ResetPassword(r.Context(), credentials.ResetCode, credentials.Email, credentials.Password); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *PasswordRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/users/password-reset", h.ResetPassword).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/password-reset/confirm", h.ResetPasswordConfirm).Methods(http.MethodPost)
}
