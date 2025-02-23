package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Servicers struct {
	db *sql.DB
}

type Handler struct {
	handlerFunc func(s *Servicers, w http.ResponseWriter, r *http.Request)
	servicers   *Servicers
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.handlerFunc(handler.servicers, w, r)
}

type User struct {
	Email    string
	Password string
}

func UserCreateHandler(s *Servicers, w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Look into exec closing requirements
	s.db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", user.Email, user.Password)
	fmt.Println(user)
}

func UserGetHandler(s *Servicers, w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "UserGetHandler!")
}

func Router(servicers *Servicers) *mux.Router {
	router := mux.NewRouter()

	userRouter := router.PathPrefix("/api/user").Subrouter()
	userRouter.Handle("/", Handler{UserCreateHandler, servicers}).Methods("POST")
	userRouter.Handle("/", Handler{UserGetHandler, servicers}).Methods("GET")

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.Handle("/", http.FileServer(http.Dir("./frontend")))
	return router
}
