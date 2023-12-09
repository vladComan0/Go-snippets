package main

import (
	"net/http"
	"net/url"
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

func TestUserSignup(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validName     = "Bob"
		validEmail    = "bob@example.com"
		validPassword = "v@lidP@$$word"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "invalidToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty e-mail",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid e-mail",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "P@$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate e-mail",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", subtest.userName)
			form.Add("email", subtest.userEmail)
			form.Add("password", subtest.userPassword)
			form.Add("csrf_token", subtest.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			assert.Equal(t, code, subtest.wantCode)

			if subtest.wantFormTag != "" {
				assert.StringContains(t, body, subtest.wantFormTag)
			}
		})
	}
}

func TestSnippetCreate(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("Unauthenticated", func(t *testing.T) {
		code, headers, _ := ts.get(t, "/snippet/create")

		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/user/login")
	})

	t.Run("Authenticated", func(t *testing.T) {
		_, _, body := ts.get(t, "/user/login")
		csrfToken := extractCSRFToken(t, body)

		// Mock a POST /user/login request using the exctracted CSRF token
		form := url.Values{}

		form.Add("email", "alice@example.com")
		form.Add("password", "pa$$word")
		form.Add("csrf_token", csrfToken)

		ts.postForm(t, "/user/login", form)

		// Check that the create snippet form is shown to the authenticated user
		code, _, body := ts.get(t, "/snippet/create")

		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, "<form action='/snippet/create' method='POST'>")
	})
}
