package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type UserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	userService services.UserService
	authService services.AuthService
	router      *mux.Router
}

func RegisterUserHandler(userService services.UserService, authService services.AuthService, router *mux.Router) {
	handler := &userHandler{
		userService: userService,
		authService: authService,
		router:      router,
	}
	handler.registerRoutes()
}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request) {
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

	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.authService.StoreRefreshToken(r.Context(), usr.ID, refreshToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request) {
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

func (h *userHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// todo
	// get refresh token from request
	// get user id from refresh token
	// return newly created jwt
}

func (h *userHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// todo
	// revoke refresh token from database
	// add option to remove all refresh tokens
}

func (h *userHandler) registerRoutes() {
	h.router.HandleFunc("/users/register", h.Register).Methods(http.MethodPost)
	h.router.HandleFunc("/users/login", h.Login).Methods(http.MethodPost)
	// router.HandleFunc("/users/refresh-token", RefreshToken).Methods("POST")
	// router.HandleFunc("/users/logout", Logout).Methods("POST")
	// router.HandleFunc("/users/profile", GetProfile).Methods("GET")
	// router.HandleFunc("/users/update-profile", UpdateProfile).Methods("POST")
	// router.HandleFunc("/users/change-password", ChangePassword).Methods("POST")
	// router.HandleFunc("/users/forgot-password", ForgotPassword).Methods("POST")
	// router.HandleFunc("/users/reset-password", ResetPassword).Methods("POST")
	// router.HandleFunc("/users", GetUsers).Methods("GET")
	// router.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE")
}
