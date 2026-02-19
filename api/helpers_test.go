package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestApplication() *application {
	return &application{
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		Redis:    nil,
	}
}

func TestWriteJSON(t *testing.T) {
	app := newTestApplication()

	tests := []struct {
		name           string
		status         int
		data           any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Write simple object",
			status:         http.StatusOK,
			data:           map[string]any{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "Write array",
			status:         http.StatusOK,
			data:           []string{"item1", "item2"},
			expectedStatus: http.StatusOK,
			expectedBody:   `["item1","item2"]`,
		},
		{
			name:           "Write with different status",
			status:         http.StatusCreated,
			data:           map[string]any{"created": true},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"created":true}`,
		},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		err := app.writeJSON(w, tt.status, tt.data)
		if err != nil {
			t.Fatalf("writeJSON() error = %v", err)
		}

		if w.Code != tt.expectedStatus {
			t.Errorf("writeJSON() status = %v, want %v", w.Code, tt.expectedStatus)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("writeJSON() Content-Type = %v, want application/json", w.Header().Get("Content-Type"))
		}

		if w.Body.String() != tt.expectedBody {
			t.Errorf("writeJSON() body = %v, want %v", w.Body.String(), tt.expectedBody)
		}
	}
}

func TestReadJSON(t *testing.T) {
	app := newTestApplication()

	tests := []struct {
		name        string
		body        string
		expected    map[string]string
		expectError bool
	}{
		{
			name:        "Valid JSON",
			body:        `{"url":"https://example.com"}`,
			expected:    map[string]string{"url": "https://example.com"},
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			body:        `{"url":invalid}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Empty JSON object",
			body:        `{}`,
			expected:    map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
		var data map[string]string
		err := app.readJSON(r, &data)

		if tt.expectError && err == nil {
			t.Error("readJSON() expected error but got none")
		}

		if !tt.expectError && err != nil {
			t.Errorf("readJSON() unexpected error = %v", err)
		}

		if !tt.expectError && tt.expected != nil {
			if len(data) != len(tt.expected) {
				t.Errorf("readJSON() data length = %v, want %v", len(data), len(tt.expected))
			}
			for k, v := range tt.expected {
				if data[k] != v {
					t.Errorf("readJSON() data[%s] = %v, want %v", k, data[k], v)
				}
			}
		}
	}
}

func TestBadRequest(t *testing.T) {
	app := newTestApplication()

	testBadRequestErr := errors.New("Invalid input")

	w := httptest.NewRecorder()
	app.badRequest(w, testBadRequestErr)

	if w.Code != http.StatusBadRequest {
		t.Errorf("badRequest() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var response map[string]any
	json.NewDecoder(w.Body).Decode(&response)

	if ok := response["error"].(bool); !ok {
		t.Errorf("badRequest() response error field = %v, want true", response["error"])
	}

	if _, ok := response["message"].(string); !ok {
		t.Error("badRequest() response should have message field")
	}
}
