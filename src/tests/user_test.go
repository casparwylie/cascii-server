package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var USER_API = "http://localhost:8000/api/user/"

func TestCreateUser_shortPassword(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	resp := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "123"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Password too short", respBody.Error)
}

func TestCreateUser_invalidEmail(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	resp := Post(
		USER_API,
		CreateUserRequest{Email: "testtest.com", Password: "12345"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Invalid email", respBody.Error)
}

func TestCreateUser_userExists(t *testing.T) {
	clearDb()
	// Create user
	Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)
	// Creater same user again
	var respBody GenericResponse
	resp := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "123456"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "User already exists", respBody.Error)
}

func TestCreateUser_success(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	resp := Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "", respBody.Error)
}

func TestAuthUser_badCredentials(t *testing.T) {
	clearDb()
	Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)
	var respBody GenericResponse
	resp := Post(
		USER_API+"auth",
		AuthUserRequest{Email: "test@test.com", Password: "123456"},
		&respBody,
	)
	cookie, err := GetCookie(resp, "sessionKey")
	assert.Error(t, err)
	assert.Nil(t, cookie)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "User not found", respBody.Error)
}

func TestAuthUser_successful(t *testing.T) {
	clearDb()
	Post(
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)
	var respBody GenericResponse
	resp := Post(
		USER_API+"auth",
		AuthUserRequest{Email: "test@test.com", Password: "12345"},
		&respBody,
	)
	cookie, err := GetCookie(resp, "sessionKey")
	assert.NoError(t, err)
	assert.NotEmpty(t, cookie)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "", respBody.Error)
}

func TestLogoutUser_successful(t *testing.T) {
	clearDb()
	client := MakeCookieClient()
	PostWithClient(
		client,
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)
	PostWithClient(
		client,
		USER_API+"auth",
		AuthUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)

	var respBody UserResponse
	resp := GetWithClient(client, USER_API+"logout", &respBody)
	cookie, _ := GetCookie(resp, "sessionKey")
	respGet := GetWithClient(client, USER_API, &respBody)

	assert.Equal(t, http.StatusUnauthorized, respGet.StatusCode)
	assert.Empty(t, cookie.Value)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUser_successful(t *testing.T) {
	clearDb()
	client := MakeCookieClient()
	PostWithClient(
		client,
		USER_API,
		CreateUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)
	PostWithClient(
		client,
		USER_API+"auth",
		AuthUserRequest{Email: "test@test.com", Password: "12345"},
		&GenericResponse{},
	)

	var respBody UserResponse
	resp := GetWithClient(client, USER_API, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "test@test.com", respBody.Email)
	assert.Equal(t, 1, respBody.Id)
}

func TestGetUser_unauthorized(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	resp := Get(USER_API, &respBody)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", respBody.Error)
}
