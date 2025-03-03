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
)

var db, err = sql.Open("mysql", fmt.Sprintf(
	"%s:%s@tcp(%s:%s)/%s",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASS"),
	os.Getenv("DB_HOST"),
	os.Getenv("DB_PORT"),
	os.Getenv("DB_NAME"),
))

func clearDb() {
	tables := []string{"sessions", "users", "drawings"}
	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
		db.Exec(fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT=1", table))
	}
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

func LoginUser() (int, *http.Client) {
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
	GetWithClient(client, USER_API, &respBody)
	return respBody.Id, client
}
