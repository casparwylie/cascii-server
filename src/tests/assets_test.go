package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var BASE_URL = "http://localhost:8000/"

func TestCoreHtml_exists(t *testing.T) {
	resp, err := http.Get(BASE_URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "// :)")
}

func TestStatic_exists(t *testing.T) {
	resp, err := http.Get(BASE_URL + "static/serverLayer.js")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "// :)")
}
