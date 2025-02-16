package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	u "github.com/dgyurics/marketplace/utilities"
)

type PasswordResetRoutes struct {
	router
	servicePasswordReset services.PasswordResetService
	serviceUser          services.UserService
	serviceEmail         services.EmailSender
}

func NewPasswordResetRoutes(servicePR services.PasswordResetService, serviceUsr services.UserService, emailSndr services.EmailSender, router router) *PasswordResetRoutes {
	return &PasswordResetRoutes{
		router:               router,
		servicePasswordReset: servicePR,
		serviceUser:          serviceUsr,
		serviceEmail:         emailSndr,
	}
}

func (h *PasswordResetRoutes) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
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
	if usr == nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "User does not exist")
		return
	}

	// Generate password reset code
	code, err := h.servicePasswordReset.GeneratePasswordResetCode(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Store password reset code
	if err := h.servicePasswordReset.StorePasswordResetCode(r.Context(), code, usr.ID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Send password reset email
	go func(recipientEmail, code string) {
		emailBody, err := u.GeneratePasswordResetEmail(recipientEmail, code)
		if err != nil {
			slog.Error("Error loading email template: ", "error", err)
			return
		}
		if err := h.serviceEmail.SendEmail([]string{recipientEmail}, "Password Reset", emailBody, true); err != nil {
			slog.Error("Error sending email: ", "error", err)
		}
	}(usr.Email, code)

	u.RespondSuccess(w)
}

func (h *PasswordResetRoutes) ResetPasswordConfirm(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
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
	valid, err := h.servicePasswordReset.ValidatePasswordResetCode(r.Context(), credentials.ResetCode, credentials.Email)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if !valid {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid reset code")
		return
	}

	// Reset the password
	if err := h.servicePasswordReset.ResetPassword(r.Context(), credentials.ResetCode, credentials.Email, credentials.Password); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *PasswordResetRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/users/password-reset", h.ResetPassword).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/password-reset/confirm", h.ResetPasswordConfirm).Methods(http.MethodPost)
}
