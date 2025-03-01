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

type CreateSessionRequest struct {
	Email    string
	Password string
}

type CrudHandler interface {
	Get(db *sql.DB, w http.ResponseWriter, r *http.Request)
	Create(db *sql.DB, w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	Servicers   *Servicers
	CrudHandler CrudHandler
}

func RequireUser(db *sql.DB, w http.ResponseWriter, r *http.Request) int {
	sessionCookie, err := r.Cookie("sessionKey")
	if err == nil {
		if userId := GetSessionUserId(db, sessionCookie.Value); userId != -1 {
			return userId
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(GenericResponse{Error: "Unauthorized"})
	return -1
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "POST":
		handler.CrudHandler.Create(handler.Servicers.db, w, r)
	case "GET":
		handler.CrudHandler.Get(handler.Servicers.db, w, r)
	}
}

type UserHandler struct{}
type SessionHandler struct{}

func (handler UserHandler) Create(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

func (handler UserHandler) Get(db *sql.DB, w http.ResponseWriter, r *http.Request) {

	userId := RequireUser(db, w, r)
	if userId == -1 {
		return
	}
	if email := GetUserById(db, userId); email != "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UserResponse{Email: email})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Error: "User not found"})
}

func (handler SessionHandler) Create(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var request CreateSessionRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if userId := Authenticate(db, request.Email, request.Password); userId != -1 {
		sessionKey := CreateSession(db, userId)
		cookie := &http.Cookie{
			Name:     "sessionKey",
			Value:    sessionKey,
			HttpOnly: true,
			Secure:   true,
		}
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(GenericResponse{Error: ""})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Error: "Login failed"})
}

func (handler SessionHandler) Get(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Error: ""})
}

func Router(servicers *Servicers) *mux.Router {
	router := mux.NewRouter()

	var userHandler = Handler{servicers, UserHandler{}}
	var sessionHandler = Handler{servicers, SessionHandler{}}

	userRouter := router.PathPrefix("/api/user").Subrouter()
	userRouter.Handle("/", userHandler)
	userRouter.Handle("/{id}", userHandler)

	sessionRouter := router.PathPrefix("/api/session").Subrouter()
	sessionRouter.Handle("/", sessionHandler)

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.Handle("/", http.FileServer(http.Dir("./frontend")))

	return router
}
