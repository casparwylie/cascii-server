package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var USER_API string = "http://localhost:8000/api/user/"
var SESSION_API string = "http://localhost:8000/api/session/"

var db, err = sql.Open("mysql", fmt.Sprintf(
	"%s:%s@tcp(%s:%s)/%s",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASS"),
	os.Getenv("DB_HOST"),
	os.Getenv("DB_PORT"),
	os.Getenv("DB_NAME"),
))

func clearDb() {
	db.Exec("TRUNCATE users; TRUNCATE sessions;")
}

func Post(url string, request any, t any) int {
	var r bytes.Buffer
	json.NewEncoder(&r).Encode(request)
	resp, err := http.Post(url, "application/json", &r)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func Get(url string, t any) int {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func TestCreateUser_shortPassword(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	status := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "123"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Password too short", respBody.Error)
}

func TestCreateUser_invalidEmail(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	status := Post(
		USER_API,
		CreateUserRequest{Email: "testtest.com", Password: "12345"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Invalid email", respBody.Error)
}

func TestCreateUser_userExists(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	// Create user
	status := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	// Creater same user again
	status = Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "123456"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "User already exists", respBody.Error)
}

func TestCreateUser_createsSuccessfully(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	status := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, "", respBody.Error)
}

func TestCreateSession_badAuth(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	status := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	status := Post(
		SESSION_API,
		CreateSessionRequest{Email: "test@test.com", Password: "123456"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Login failed", respBody.Error)
}

func TestCreateSession_createsSuccessfully(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	status := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		GenericResponse{},
	)
	status := Post(
		SESSION_API,
		CreateSessionRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	// TODO assert cookie creation
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, "", respBody.Error)
}

func TestGetUser_Unauthorized(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	status := Get(USER_API+"999", &respBody)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Equal(t, "Unauthorized", respBody.Error)
}

func TestGetUser_getsSuccessfully(t *testing.T) {
	clearDb()
	status := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		GenericResponse{},
	)
	status := Post(
		SESSION_API,
		CreateSessionRequest{Email: "test@test.com", Password: "12345"},
		GenericResponse{},
	)
	// TODO send cookie
	status := Get(USER_API, &respBody)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, "", respBody.Error)
}
