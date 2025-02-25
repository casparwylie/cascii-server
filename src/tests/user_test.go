package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestCreateUser_shortPassword(t *testing.T) {
	var respBody GenericResponse
	status := Post(
		"http://localhost:8000/api/user/",
		CreateUserRequest{Email: "test@test.com", Password: "123"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Password too short", respBody.Error)
}

func TestCreateUser_invalidEmail(t *testing.T) {
	var respBody GenericResponse
	status := Post(
		"http://localhost:8000/api/user/",
		CreateUserRequest{Email: "testtest.com", Password: "12345"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Invalid email", respBody.Error)
}

func TestCreateUser_userExists(t *testing.T) {
	var respBody GenericResponse
	// Creater user
	status := Post(
		"http://localhost:8000/api/user/",
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	// Creater same user again
	status = Post(
		"http://localhost:8000/api/user/",
		CreateUserRequest{Email: "test@test.com", Password: "123456"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "User already exists", respBody.Error)
}

func TestCreateUser_createsSuccessfully(t *testing.T) {
	var respBody GenericResponse
	status := Post(
		"http://localhost:8000/api/user/",
		CreateUserRequest{Email: "test1@test.com", Password: "12345"},
		&respBody,
	)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, "", respBody.Error)
}
