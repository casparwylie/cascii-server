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
	tables := []string{"sessions", "users", "mutable_drawings", "immutable_drawings"}
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

func Post(url string, req any, res any) *http.Response {
	return PostWithClient(&http.Client{}, url, req, res)
}

func Patch(url string, req any, res any) *http.Response {
	return PatchWithClient(&http.Client{}, url, req, res)
}

func Get(url string, res any) *http.Response {
	return GetWithClient(&http.Client{}, url, res)
}

func Delete(url string, res any) *http.Response {
	return DeleteWithClient(&http.Client{}, url, res)
}

func PatchWithClient(client *http.Client, url string, req any, res any) *http.Response {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(req)                            // Writes to buf
	request, err := http.NewRequest(http.MethodPatch, url, &buf) // Reads from buf
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		panic(err)
	}
	return resp
}

func PostWithClient(client *http.Client, url string, req any, res any) *http.Response {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(req)                       // Writes to buf
	resp, err := client.Post(url, "application/json", &buf) // Reads from buf
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		panic(err)
	}
	return resp
}

func GetWithClient(client *http.Client, url string, res any) *http.Response {
	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		panic(err)
	}
	return resp
}

func DeleteWithClient(client *http.Client, url string, res any) *http.Response {
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		panic(err)
	}
	return resp
}

func LoginUser(email string) (int, *http.Client) {
	client := MakeCookieClient()
	PostWithClient(
		client,
		USER_API,
		CreateUserRequest{Email: email, Password: "12345"},
		&GenericResponse{},
	)
	PostWithClient(
		client,
		USER_API+"auth",
		AuthUserRequest{Email: email, Password: "12345"},
		&GenericResponse{},
	)

	var respBody UserResponse
	GetWithClient(client, USER_API, &respBody)
	return respBody.Id, client
}
