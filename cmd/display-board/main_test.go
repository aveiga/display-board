package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandleHome(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleHome)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check if the response contains expected HTML elements
	expectedStrings := []string{
		"<title>Display Board</title>",
		"<h1>Display Board</h1>",
		"<table>",
		"<thead>",
	}

	for _, str := range expectedStrings {
		if !strings.Contains(rr.Body.String(), str) {
			t.Errorf("handler response missing %s", str)
		}
	}
}

func TestHandleMessageCreationPage(t *testing.T) {
	req, err := http.NewRequest("GET", "/messages", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleMessageCreationPage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedStrings := []string{
		"<title>Display Board - Create Message</title>",
		"<h1>Create New Message</h1>",
		`<form hx-post="/message"`,
		"<textarea",
	}

	for _, str := range expectedStrings {
		if !strings.Contains(rr.Body.String(), str) {
			t.Errorf("handler response missing %s", str)
		}
	}
}

func TestHandleMessageSubmission(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		formData     url.Values
		expectedCode int
	}{
		{
			name:         "Valid POST request",
			method:       "POST",
			formData:     url.Values{"message": {"Test message"}},
			expectedCode: http.StatusSeeOther,
		},
		{
			name:         "Invalid method",
			method:       "GET",
			formData:     nil,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Empty message",
			method:       "POST",
			formData:     url.Values{"message": {""}},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.formData != nil {
				req, err = http.NewRequest(tt.method, "/message", strings.NewReader(tt.formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, err = http.NewRequest(tt.method, "/message", nil)
			}

			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleMessageSubmission)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}
		})
	}
}
