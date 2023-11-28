package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vladComan0/go-snippets/internal/assert"
)

func TestSecureHeaders(t *testing.T) {
	responseRecorder := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP handler that we can pass to our secureHeaders middleware
	// which writes a 200 status code and an "OK" response body.
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Pass the mock HTTP handler to our secureHeaders middleware.
	secureHeaders(mockHandler).ServeHTTP(responseRecorder, request)

	responseResult := responseRecorder.Result()

	headerTests := []struct {
		name     string
		expected string
	}{
		{
			name:     "Content-Security-Policy",
			expected: "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		},
		{
			name:     "Referrer-Policy",
			expected: "origin-when-cross-origin",
		},
		{
			name:     "X-Content-Type-Options",
			expected: "nosniff",
		},
		{
			name:     "X-Frame-Options",
			expected: "deny",
		},
		{
			name:     "X-XSS-Protection",
			expected: "0",
		},
	}

	for _, subtest := range headerTests {
		t.Run(subtest.name, func(t *testing.T) {
			assert.Equal(t, responseResult.Header.Get(subtest.name), subtest.expected)
		})
	}
	// // Check that the middleware has correctly set the Content-Security-Policy
	// // header on the response.
	// expectedValue := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	// assert.Equal(t, responseResult.Header.Get("Content-Security-Policy"), expectedValue)

	// // Check that the middleware has correctly set the Referrer-Policy
	// // header on the response.
	// expectedValue = "origin-when-cross-origin"
	// assert.Equal(t, responseResult.Header.Get("Referrer-Policy"), expectedValue)

	// // Check that the middleware has correctly set the X-Content-Type-Options
	// // header on the response.
	// expectedValue = "nosniff"
	// assert.Equal(t, responseResult.Header.Get("X-Content-Type-Options"), expectedValue)

	// // Check that the middleware has correctly set the X-Frame-Options
	// // header on the response.
	// expectedValue = "deny"
	// assert.Equal(t, responseResult.Header.Get("X-Frame-Options"), expectedValue)

	// // Check that the middleware has correctly set the X-XSS-Protection
	// // header on the response.
	// expectedValue = "0"
	// assert.Equal(t, responseResult.Header.Get("X-XSS-Protection"), expectedValue)

	// Check that the middleware has correctly called the next handler in line
	// and the response status code and body are as expected.
	assert.Equal(t, responseResult.StatusCode, http.StatusOK)

	defer responseResult.Body.Close()
	responseBody, err := io.ReadAll(responseResult.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(responseBody)
	assert.Equal(t, string(responseBody), "OK")
}
