package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vladComan0/letsgo/internal/assert"
)

func TestPing(t *testing.T) {
	responseRecorder := httptest.NewRecorder()

	// Initialize a new dummy http.Request.
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	ping(responseRecorder, request)

	// Check that the HTTP status code is 200.
	responseResult := responseRecorder.Result()
	defer responseResult.Body.Close()
	assert.Equal(t, responseResult.StatusCode, http.StatusOK)

	// Check that the HTTP response body is "OK".
	responseBody, err := io.ReadAll(responseResult.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(responseBody)
	assert.Equal(t, string(responseBody), "OK")
}
