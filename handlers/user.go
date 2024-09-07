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
	router      *mux.Router
}

func RegisterUserHandler(userService services.UserService, router *mux.Router) {
	handler := &userHandler{
		userService: userService,
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
	// todo
	// get jwt and refresh token from user service
	// return jwt and refresh token in response
	jwt := "jwt"
	refreshToken := "refreshToken"
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":         jwt,
		"refresh_token": refreshToken,
	})
}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	usr := models.User{
		Email:    credentials.Email,
		Phone:    credentials.Phone,
		Password: credentials.Password,
	}
	err := h.userService.VerifyCredentials(r.Context(), &usr)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// todo
	// get jwt and refresh token from user service
	// return jwt and refresh token in response
	jwt := "jwt"
	refreshToken := "refreshToken"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token":         jwt,
		"refresh_token": refreshToken,
	})
}

func (h *userHandler) registerRoutes() {
	h.router.HandleFunc("/users/register", h.Register).Methods("POST")
	h.router.HandleFunc("/users/login", h.Login).Methods("POST")
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
