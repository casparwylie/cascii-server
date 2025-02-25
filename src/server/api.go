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
	email string
}

type CreateUserRequest struct {
	Email    string
	Password string
}

type UserHandler struct {
	servicers *Servicers
}

func (handler UserHandler) create(w http.ResponseWriter, r *http.Request) {
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
	if UserExists(handler.servicers.db, request.Email) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GenericResponse{Error: "User already exists"})
		return
	}
	if success := CreateUser(handler.servicers.db, request.Email, request.Password); success != true {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenericResponse{Error: "Unknown error"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GenericResponse{Error: ""})
}

func (handler UserHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := GetUserById(handler.servicers.db, vars["id"])
	if email == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserResponse{email: email})
}

func (handler UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "POST":
		handler.create(w, r)
	case "GET":
		handler.get(w, r)
	}
}
func Router(servicers *Servicers) *mux.Router {
	router := mux.NewRouter()

	userRouter := router.PathPrefix("/api/user").Subrouter()
	userRouter.Handle("/", UserHandler{servicers}).Methods("POST")
	userRouter.Handle("/{id}", UserHandler{servicers}).Methods("GET")

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.Handle("/", http.FileServer(http.Dir("./frontend")))
	return router
}
