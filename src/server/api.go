package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

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

type User struct {
	Email    string
	Password string
}

type UserHandler struct {
	servicers *Servicers
}

func (handler UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	errorMessage := CreateUser(handler.servicers.db, user.Email, user.Password)
	if errorMessage == "" {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(GenericResponse{Error: errorMessage})
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
