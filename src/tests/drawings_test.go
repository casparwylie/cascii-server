package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var DRAWINGS_API = "http://localhost:8000/api/drawings/"

type BadData struct {
	Test string
}

func TestCreateImmutableDrawing_succesful(t *testing.T) {
	clearDb()
	var respBody CreateImmutableDrawingResponse
	resp := Post(
		DRAWINGS_API+"immutable",
		CreateImmutableDrawingRequest{Data: "{\"test\": \"test\"}"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "c59e4b", respBody.ShortKey)
}

func TestGetImmutableDrawing_successful(t *testing.T) {
	clearDb()
	var respBody1 CreateImmutableDrawingResponse
	resp := Post(
		DRAWINGS_API+"immutable",
		CreateImmutableDrawingRequest{Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	var respBody2 GetImmutableDrawingResponse
	resp = Get(
		DRAWINGS_API+"immutable/"+respBody1.ShortKey,
		&respBody2,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"test\": \"test\"}", respBody2.Data)
	assert.NotEmpty(t, respBody2.CreatedAt)
}

func TestGetImmutableDrawing_notFound(t *testing.T) {
	clearDb()
	resp := Get(
		DRAWINGS_API+"immutable/999",
		&GetImmutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateMutableDrawing_unauthorized(t *testing.T) {
	clearDb()
	var respBody CreateMutableDrawingResponse
	resp := Post(
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody,
	)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateMutableDrawing_badRequest(t *testing.T) {
	clearDb()
	_, client := LoginUser()
	resp := PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		BadData{Test: "test"},
		&CreateMutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateMutableDrawing_success(t *testing.T) {
	clearDb()
	_, client := LoginUser()
	var respBody CreateMutableDrawingResponse
	resp := PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody,
	)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, 1, respBody.Id)
}
