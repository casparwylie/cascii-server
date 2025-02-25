package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	values := map[string]string{"email": "test@test.com", "password": "test123"}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post("http://localhost:8000/api/user/", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		// handle error
		panic(err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	t.Fatalf("first trest")
}
