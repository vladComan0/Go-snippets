package main

import (
	"net/http"
	"testing"

	"github.com/vladComan0/go-snippets/internal/assert"
)

func TestPing(t *testing.T) {
	// Create a new instance of the application struct.
	// The reason we only set the loggers from the app struct is that
	// these are needed by the logRequest and recoverPanic middlewares, which
	// are used by the application on every route.
	app := newTestApplication(t)

	// Create a new test server, passing in the value returned by
	// app.routes() method as the handler for the server.
	// This starts an HTTPS server which listens on a randomly-chosen
	// port on the local machine for the duration of the test
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	statusCode, _, responseBody := ts.get(t, "/ping")

	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, responseBody, "OK")
}

func TestSnippetView(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/snippet/view/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/snippet/view/1.34",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/snippet/view/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/snippet/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.name, func(t *testing.T) {
			code, _, body := ts.get(t, subtest.urlPath)

			assert.Equal(t, code, subtest.wantCode)
			if subtest.wantBody != "" {
				assert.StringContains(t, body, subtest.wantBody)
			}
		})
	}
}
