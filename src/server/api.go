package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Servicers struct {
	db *sql.DB
}

type GenericResponse struct {
	Error string `json:"error"`
}

type UserResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateImmutableDrawingRequest struct {
	Data string `json:"data"`
}

type CreateImmutableDrawingResponse struct {
	ShortKey string `json:"short_key"`
}

type GetImmutableDrawingResponse struct {
	Data      string `json:"data"`
	CreatedAt string `json:"created_at"`
}

type CreateMutableDrawingRequest struct {
	Data string `json:"data"`
	Name string `json:"name"`
}

type CreateMutableDrawingResponse struct {
	Id int `json:"id"`
}

type UpdateMutableDrawingRequest struct {
	Id   int    `json:"id"`
	Data string `json:"data"`
	Name string `json:"name"`
}

type GetMutableDrawingResponse struct {
	Id        int    `json:"id"`
	UserId    int    `json:"user_id"`
	Data      string `json:"data"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type AuthHandler struct {
	Servicers   *Servicers
	HandlerFunc func(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	Servicers   *Servicers
	HandlerFunc func(db *sql.DB, w http.ResponseWriter, r *http.Request)
}

func WriteUnknownError(w http.ResponseWriter, err error) {
	log.Fatal(err)
	WriteGenericResponse(w, http.StatusInternalServerError, "Unknown error")
}

func WriteGenericResponse(w http.ResponseWriter, status int, err string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(GenericResponse{Error: err})
}

func WriteStructuredResponse(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func DecodeRequest(request any, w http.ResponseWriter, r *http.Request) bool {
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteGenericResponse(w, http.StatusBadRequest, "Bad request")
	}
	return err == nil
}

func (handler AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionCookie, err := r.Cookie("sessionKey")
	if err != nil {
		WriteGenericResponse(w, http.StatusUnauthorized, "Unauthorized")
	}
	userId, err := GetSessionUserId(handler.Servicers.db, sessionCookie.Value)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	if userId > -1 {
		handler.HandlerFunc(handler.Servicers.db, userId, w, r)
		return
	}
	WriteGenericResponse(w, http.StatusUnauthorized, "Unauthorized")
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	handler.HandlerFunc(handler.Servicers.db, w, r)
}

func CreateUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequest
	if !DecodeRequest(&request, w, r) {
		return
	}
	if len(request.Password) < 5 {
		WriteGenericResponse(w, http.StatusOK, "Password too short")
		return
	}
	parsedEmail, err := mail.ParseAddress(request.Email)
	if err != nil || parsedEmail.Address != request.Email {
		WriteGenericResponse(w, http.StatusOK, "Invalid email")
		return
	}
	exists, err := UserExists(db, request.Email)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	if exists {
		WriteGenericResponse(w, http.StatusOK, "User already exists")
		return
	}
	if err := CreateUser(db, request.Email, request.Password); err != nil {
		WriteUnknownError(w, err)
		return
	}
	WriteGenericResponse(w, http.StatusCreated, "")
}

func GetUserHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
	email, err := GetUserById(db, userId)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	if email == "" {
		WriteGenericResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteStructuredResponse(w, http.StatusOK, UserResponse{Id: userId, Email: email})
}

func AuthUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var request AuthUserRequest
	if !DecodeRequest(&request, w, r) {
		return
	}
	userId, err := Authenticate(db, request.Email, request.Password)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	if userId == -1 {
		WriteGenericResponse(w, http.StatusOK, "User not found")
		return
	}
	sessionKey, err := CreateSession(db, userId)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	cookie := &http.Cookie{
		Name:     "sessionKey",
		Value:    sessionKey,
		HttpOnly: true,
		Path:     "/",
		// TODO: Secure?
	}
	http.SetCookie(w, cookie)
	WriteGenericResponse(w, http.StatusAccepted, "")
}

func LogoutUserHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
	if err := DeleteSession(db, userId); err != nil {
		WriteUnknownError(w, err)
		return
	}
	cookie := &http.Cookie{
		Name:     "sessionKey",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
	WriteGenericResponse(w, http.StatusOK, "")
}

func CreateImmutableDrawingHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var request CreateImmutableDrawingRequest
	if !DecodeRequest(&request, w, r) {
		return
	}
	shortKey, err := CreateImmutableDrawing(db, request.Data)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	WriteStructuredResponse(w, http.StatusOK, CreateImmutableDrawingResponse{ShortKey: shortKey})
}

func GetImmutableDrawingHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	shortKey := mux.Vars(r)["short_key"]
	data, createdAt, err := GetImmutableDrawing(db, shortKey)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	if data == "" {
		WriteGenericResponse(w, http.StatusNotFound, "Drawing not found")
		return
	}
	WriteStructuredResponse(w, http.StatusOK, GetImmutableDrawingResponse{Data: data, CreatedAt: createdAt})
}

func CreateMutableDrawingHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
	var request CreateMutableDrawingRequest
	if !DecodeRequest(&request, w, r) {
		return
	}
	id, err := CreateMutableDrawing(db, request.Data, request.Name, userId)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	WriteStructuredResponse(w, http.StatusCreated, CreateMutableDrawingResponse{Id: id})
}

func GetMutableDrawingHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteGenericResponse(w, http.StatusBadRequest, "Bad request")
		return
	}
	userId, name, data, createdAt, err := GetMutableDrawing(db, id)
	if err != nil {
		WriteUnknownError(w, err)
		return
	}
	if name == "" {
		WriteGenericResponse(w, http.StatusNotFound, "Drawing not found")
		return
	}
	WriteStructuredResponse(w, http.StatusOK, GetMutableDrawingResponse{
		Id: id, UserId: userId, Data: data, Name: name, CreatedAt: createdAt,
	})
}

func UpdateMutableDrawingHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
	var request UpdateMutableDrawingRequest
	if !DecodeRequest(&request, w, r) {
		return
	}
	if err := UpdateMutableDrawing(db, request.Id, request.Data, request.Name, userId); err != nil {
		WriteUnknownError(w, err)
		return
	}
	WriteGenericResponse(w, http.StatusOK, "")
}

func DeleteMutableDrawingHandler(db *sql.DB, userId int, w http.ResponseWriter, r *http.Request) {
}

func Router(servicers *Servicers) *mux.Router {
	router := mux.NewRouter()

	userRouter := router.PathPrefix("/api/user").Subrouter()
	userRouter.Handle("/", Handler{servicers, CreateUserHandler}).Methods("POST")
	userRouter.Handle("/", AuthHandler{servicers, GetUserHandler}).Methods("GET")
	userRouter.Handle("/auth", Handler{servicers, AuthUserHandler}).Methods("POST")
	userRouter.Handle("/logout", AuthHandler{servicers, LogoutUserHandler}).Methods("GET")

	drawingsRouter := router.PathPrefix("/api/drawings").Subrouter()
	drawingsRouter.Handle("/immutable", Handler{servicers, CreateImmutableDrawingHandler}).Methods("POST")
	drawingsRouter.Handle("/immutable/{short_key}", Handler{servicers, GetImmutableDrawingHandler}).Methods("GET")
	drawingsRouter.Handle("/mutable", AuthHandler{servicers, CreateMutableDrawingHandler}).Methods("POST")
	drawingsRouter.Handle("/mutable", AuthHandler{servicers, UpdateMutableDrawingHandler}).Methods("PATCH")
	drawingsRouter.Handle("/mutable", AuthHandler{servicers, GetMutableDrawingHandler}).Methods("DELETE")
	drawingsRouter.Handle("/mutable/{id}", AuthHandler{servicers, DeleteMutableDrawingHandler}).Methods("GET")

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.Handle("/", http.FileServer(http.Dir("./frontend")))

	return router
}
