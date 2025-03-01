package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/mail"

	"github.com/gorilla/mux"
)

type Servicers struct {
	db *sql.DB
}

type GenericResponse struct {
	Error string `json:"error"`
}

type UserResponse struct {
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Email    string
	Password string
}

type AuthUserRequest struct {
	Email    string
	Password string
}

type AuthHandler struct {
	Servicers   *Servicers
	HandlerFunc func(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	Servicers   *Servicers
	HandlerFunc func(db *sql.DB, w http.ResponseWriter, r *http.Request)
}

func (handler AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionCookie, err := r.Cookie("sessionKey")
	if err == nil {
		if userId := GetSessionUserId(handler.Servicers.db, sessionCookie.Value); userId != -1 {
			handler.HandlerFunc(handler.Servicers.db, userId, w, r)
			return
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(GenericResponse{Error: "Unauthorized"})
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	handler.HandlerFunc(handler.Servicers.db, w, r)
}

func CreateUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(request.Password) < 5 {
		json.NewEncoder(w).Encode(GenericResponse{Error: "Password too short"})
		return
	}
	parsedEmail, err := mail.ParseAddress(request.Email)
	if err != nil || parsedEmail.Address != request.Email {
		json.NewEncoder(w).Encode(GenericResponse{Error: "Invalid email"})
		return
	}
	if UserExists(db, request.Email) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GenericResponse{Error: "User already exists"})
		return
	}
	if success := CreateUser(db, request.Email, request.Password); success != true {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Error: "Unknown error"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GenericResponse{Error: ""})
}

func GetUserHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
	if email := GetUserById(db, userId); email != "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UserResponse{Email: email})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Error: "User not found"})
}

func AuthUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var request AuthUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if userId := Authenticate(db, request.Email, request.Password); userId != -1 {
		cookie := &http.Cookie{
			Name:     "sessionKey",
			Value:    CreateSession(db, userId),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(GenericResponse{Error: ""})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Error: "User not found"})
}

func Router(servicers *Servicers) *mux.Router {
	router := mux.NewRouter()

	userRouter := router.PathPrefix("/api/user").Subrouter()
	userRouter.Handle("/", Handler{servicers, CreateUserHandler}).Methods("POST")
	userRouter.Handle("/", AuthHandler{servicers, GetUserHandler}).Methods("GET")
	userRouter.Handle("/auth", Handler{servicers, AuthUserHandler}).Methods("POST")

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.Handle("/", http.FileServer(http.Dir("./frontend")))

	return router
}
