package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var USER_API string = "http://localhost:8000/api/user/"

var db, err = sql.Open("mysql", fmt.Sprintf(
	"%s:%s@tcp(%s:%s)/%s",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASS"),
	os.Getenv("DB_HOST"),
	os.Getenv("DB_PORT"),
	os.Getenv("DB_NAME"),
))

func clearDb() {
	db.Exec("DELETE FROM sessions")
	db.Exec("DELETE FROM users")
}

func MakeCookieClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &http.Client{Jar: jar}
}

func GetCookie(resp *http.Response, name string) (*http.Cookie, error) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return nil, errors.New("No cookie found")
}

func Post(url string, request any, t any) *http.Response {
	return PostWithClient(&http.Client{}, url, request, t)
}

func Get(url string, t any) *http.Response {
	return GetWithClient(&http.Client{}, url, t)
}

func PostWithClient(client *http.Client, url string, request any, t any) *http.Response {
	var r bytes.Buffer
	json.NewEncoder(&r).Encode(request)
	resp, err := client.Post(url, "application/json", &r)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		panic(err)
	}
	return resp
}

func GetWithClient(client *http.Client, url string, t any) *http.Response {
	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		panic(err)
	}
	return resp
}

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

func TestAuthUser_success(t *testing.T) {
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

func TestGetUser_success(t *testing.T) {
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
}

func TestGetUser_Unauthorized(t *testing.T) {
	clearDb()
	var respBody GenericResponse
	resp := Get(USER_API, &respBody)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "Unauthorized", respBody.Error)
}
