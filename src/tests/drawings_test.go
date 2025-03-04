package main

import (
	"fmt"
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
	assert.Equal(t, "c59e4", respBody.ShortKey)
}

func TestCreateImmutableDrawing_duplicate(t *testing.T) {
	clearDb()
	Post(
		DRAWINGS_API+"immutable",
		CreateImmutableDrawingRequest{Data: "{\"test\": \"test\"}"},
		&CreateImmutableDrawingResponse{},
	)
	var respBody CreateImmutableDrawingResponse
	resp := Post(
		DRAWINGS_API+"immutable",
		CreateImmutableDrawingRequest{Data: "{\"test\": \"test\"}"},
		&respBody,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "c59e4", respBody.ShortKey)
}

func TestCreateImmutableDrawing_realConflict(t *testing.T) {
	/*
		The following function can be used to find two different
		values which have the same x first characters in their
		respective Hash output.

		func main() {
			template1 := "{\"test\": \"%d\"}"
			template2 := "{\"%d\": \"test\"}"
			x := 5

			i := 0
			for {
				val1 := fmt.Sprintf(template1, i)
				val2 := fmt.Sprintf(template2, i)
				if Hash(val1)[:x] == Hash(val2)[:x] {
					fmt.Println(val1)
					fmt.Println(Hash(val1))
					fmt.Println(val2)
					fmt.Println(Hash(val2))
					break
				}
				i += 1
			}
		}
	*/
	clearDb()
	var respBody1 CreateImmutableDrawingResponse
	// The following json objects are known to share the same
	// first 5 hash characters for SHA512.
	resp1 := Post(
		DRAWINGS_API+"immutable",
		CreateImmutableDrawingRequest{Data: "{\"test\": \"1798285\"}"},
		&respBody1,
	)
	var respBody2 CreateImmutableDrawingResponse
	resp2 := Post(
		DRAWINGS_API+"immutable",
		CreateImmutableDrawingRequest{Data: "{\"1798285\": \"test\"}"},
		&respBody2,
	)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	assert.Equal(t, "66d57", respBody1.ShortKey)
	assert.Equal(t, "66d57e", respBody2.ShortKey)
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
	_, client := LoginUser("test@test.com")
	resp := PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		BadData{Test: "test"},
		&CreateMutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateMutableDrawing_successful(t *testing.T) {
	clearDb()
	_, client := LoginUser("test@test.com")
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

func TestGetMutableDrawing_notFound(t *testing.T) {
	clearDb()
	_, client := LoginUser("test@test.com")
	resp := GetWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", 999),
		&GetMutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetMutableDrawing_differentUserNoAccess(t *testing.T) {
	clearDb()
	_, client1 := LoginUser("test@test.com")
	_, client2 := LoginUser("test1@test.com")
	var respBody1 CreateMutableDrawingResponse
	PostWithClient(
		client1,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	resp1 := GetWithClient(
		client1,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&GetMutableDrawingResponse{},
	)
	resp2 := GetWithClient(
		client2,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&GetMutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestGetMutableDrawing_successful(t *testing.T) {
	clearDb()
	userId, client := LoginUser("test@test.com")
	var respBody1 CreateMutableDrawingResponse
	PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	var respBody2 GetMutableDrawingResponse
	resp := GetWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&respBody2,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, respBody1.Id, respBody2.Id)
	assert.Equal(t, userId, respBody2.UserId)
	assert.Equal(t, "{\"test\": \"test\"}", respBody2.Data)
	assert.Equal(t, "test", respBody2.Name)
	assert.NotEmpty(t, respBody2.CreatedAt)
}

func TestUpdateMutableDrawing_notFound(t *testing.T) {
	clearDb()
	_, client := LoginUser("test@test.com")
	resp := PutWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", 999),
		UpdateMutableDrawingRequest{Name: "test updated", Data: "{\"test\": \"updated\"}"},
		&GenericResponse{},
	)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateMutableDrawing_successful(t *testing.T) {
	clearDb()
	userId, client := LoginUser("test@test.com")
	var respBody1 CreateMutableDrawingResponse
	PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	resp := PutWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		UpdateMutableDrawingRequest{Name: "updated", Data: "{\"test\": \"updated\"}"},
		&GenericResponse{},
	)
	var respBody2 GetMutableDrawingResponse
	GetWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&respBody2,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, respBody1.Id, respBody2.Id)
	assert.Equal(t, userId, respBody2.UserId)
	assert.Equal(t, "{\"test\": \"updated\"}", respBody2.Data)
	assert.Equal(t, "updated", respBody2.Name)
	assert.NotEmpty(t, respBody2.CreatedAt)
}

func TestUpdateMutableDrawing_differentUserNoAccess(t *testing.T) {
	clearDb()
	_, client1 := LoginUser("test@test.com")
	_, client2 := LoginUser("test1@test.com")
	var respBody1 CreateMutableDrawingResponse
	PostWithClient(
		client1,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	resp1 := PutWithClient(
		client1,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		UpdateMutableDrawingRequest{Name: "test updated", Data: "{\"test\": \"updated\"}"},
		&GenericResponse{},
	)
	resp2 := PutWithClient(
		client2,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		UpdateMutableDrawingRequest{Name: "updated again", Data: "{\"test\": \"updated again\"}"},
		&GenericResponse{},
	)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestDeleteMutableDrawing_notFound(t *testing.T) {
	clearDb()
	_, client := LoginUser("test@test.com")
	resp := DeleteWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", 999),
		&GetMutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteMutableDrawing_successful(t *testing.T) {
	clearDb()
	_, client := LoginUser("test@test.com")
	var respBody1 CreateMutableDrawingResponse
	PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	resp1 := DeleteWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&GenericResponse{},
	)
	resp2 := GetWithClient(
		client,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&GetMutableDrawingResponse{},
	)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestDeleteMutableDrawing_differentUserNoAccess(t *testing.T) {
	clearDb()
	_, client1 := LoginUser("test@test.com")
	_, client2 := LoginUser("test1@est.com")
	var respBody1 CreateMutableDrawingResponse
	PostWithClient(
		client1,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test", Data: "{\"test\": \"test\"}"},
		&respBody1,
	)
	resp1 := GetWithClient( // Don't delete as correct user otherwise the 404 won't be legit
		client1,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&GenericResponse{},
	)
	resp2 := DeleteWithClient(
		client2,
		DRAWINGS_API+fmt.Sprintf("mutable/%d", respBody1.Id),
		&GenericResponse{},
	)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestListMutableDrawings_successful(t *testing.T) {
	clearDb()
	_, client := LoginUser("test@test.com")
	PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test1", Data: "{\"test\": \"test\"}"},
		&CreateMutableDrawingResponse{},
	)
	PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test2", Data: "{\"test\": \"test\"}"},
		&CreateMutableDrawingResponse{},
	)
	PostWithClient(
		client,
		DRAWINGS_API+"mutable",
		CreateMutableDrawingRequest{Name: "test3", Data: "{\"test\": \"test\"}"},
		&CreateMutableDrawingResponse{},
	)
	var respBody ListMutableDrawingsResponse
	resp := GetWithClient(
		client,
		DRAWINGS_API+"mutables",
		&respBody,
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Results, 3)
	assert.Equal(t, "test1", respBody.Results[0].Name)
	assert.Equal(t, "test2", respBody.Results[1].Name)
	assert.Equal(t, "test3", respBody.Results[2].Name)
}
