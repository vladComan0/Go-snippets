package main

import (
	"net/http"
	"testing"

	"github.com/vladComan0/letsgo/internal/assert"
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
