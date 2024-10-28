package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpHandler(t *testing.T) {
	// Simple GET test
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	httpHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, w.Code)
	}
	var response HTTPResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Method != "GET" {
		t.Errorf("Expected method GET, got %v", response.Method)
	}

	// More robust POST test
	data := []byte(`{"name": "John Doe", "email": "john.doe@example.com"}`)
	req = httptest.NewRequest(http.MethodPost, "/test?name=john&age=40", bytes.NewBuffer(data))
	w = httptest.NewRecorder()
	httpHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Method != "POST" {
		t.Errorf("Expected method POST, got %v", response.Method)
	}
	if response.Params["age"] != "40" {
		t.Errorf("Expected age 40, got %v", response.Params["age"])
	}
	jsonResp, _ := json.Marshal(response.Json)
	if !strings.Contains(string(jsonResp), "John Doe") {
		t.Errorf("Expected John Doe in json string '%v'", string(jsonResp))
	}
}
