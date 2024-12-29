package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
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

func (h *UserRoutes) Register(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if (credentials.Email == "" && credentials.Phone == "") || credentials.Password == "" {
		http.Error(w, "Email or phone and password are required", http.StatusBadRequest)
		return
	}

	usr := models.User{
		Email:    credentials.Email,
		Phone:    credentials.Phone,
		Password: credentials.Password,
	}
	if err := h.userService.CreateUser(r.Context(), &usr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accessToken, err := h.authService.GenerateAccessToken(usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var token string
	token, err = h.authService.GenerateRefreshToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.authService.StoreRefreshToken(r.Context(), usr.ID, token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": token,
	})
}

func (h *UserRoutes) Login(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if (credentials.Email == "" && credentials.Phone == "") || credentials.Password == "" {
		http.Error(w, "Email or phone and password are required", http.StatusBadRequest)
		return
	}

	// Verify user credentials
	usr, err := h.userService.AuthenticateUser(r.Context(), &credentials)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate access token
	accessToken, err := h.authService.GenerateAccessToken(*usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store refresh token
	if err := h.authService.StoreRefreshToken(r.Context(), usr.ID, refreshToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken generates a new access token using a valid refresh token
func (h *UserRoutes) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if requestBody.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Validate the refresh token
	user, err := h.authService.ValidateRefreshToken(r.Context(), requestBody.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// TODO cycle refresh tokens
	// revoke the just validated refresh token and generate a new one

	// Generate a new access token
	accessToken, err := h.authService.GenerateAccessToken(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": requestBody.RefreshToken,
	})
}

func (h *UserRoutes) Logout(w http.ResponseWriter, r *http.Request) {
	// todo
	// revoke refresh token from database
	// add option to remove all refresh tokens
}

func (h *UserRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/users/register", h.Register).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/login", h.Login).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/users/refresh-token", h.RefreshToken).Methods("POST")
	// router.HandleFunc("/users/logout", Logout).Methods("POST")
	// router.HandleFunc("/users/profile", GetProfile).Methods("GET")
	// router.HandleFunc("/users/update-profile", UpdateProfile).Methods("POST")
	// router.HandleFunc("/users/change-password", ChangePassword).Methods("POST")
	// router.HandleFunc("/users/forgot-password", ForgotPassword).Methods("POST")
	// router.HandleFunc("/users/reset-password", ResetPassword).Methods("POST")
	// router.HandleFunc("/users", GetUsers).Methods("GET")
	// router.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE")
}
