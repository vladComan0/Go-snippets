package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/vladComan0/go-snippets/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Convert the ui.Files embedded filesystem into an http.FileServer to satisfy the http.FileSystem interface
	fileServer := http.FileServer(http.FS(ui.Files))

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.clientError(w, http.StatusNotFound)
	})

	// CSS server route
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/ping", ping)
	// CRUD + Authentication routes

	// Create a new middleware chain containing the middleware specific to our dynamic application routes.
	dynamicChain := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// Unprotected (with respect to authotization) application routes that use the "dynamic" middleware chain.
	router.Handler(http.MethodGet, "/", dynamicChain.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamicChain.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/about", dynamicChain.ThenFunc(app.about))

	router.Handler(http.MethodGet, "/user/signup", dynamicChain.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamicChain.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamicChain.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamicChain.ThenFunc(app.userLoginPost))

	// Create a new middleware chain containing the middleware specific to our dynamic application routes
	// AND the "requireAuthentication" middleware.
	protectedChain := dynamicChain.Append(app.requireAuthentication)

	// Protected (with respect to authorization) application routes that use the protected middleware chain.
	router.Handler(http.MethodGet, "/snippet/create", protectedChain.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protectedChain.ThenFunc(app.snippetCreatePost))

	router.Handler(http.MethodPost, "/user/logout", protectedChain.ThenFunc(app.userLogoutPost))

	// Create a middlware chain containing the standard middleware
	// which are to be used for every request our application receives.
	standardChain := alice.New(app.recoverPanic, app.logRequests, secureHeaders)

	return standardChain.Then(router)
}
