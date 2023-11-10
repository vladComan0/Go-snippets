package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.clientError(w, http.StatusNotFound)
	})

	// Create a new middleware chain containing the middleware specific to our dynamic application routes.
	dynamicChain := alice.New(app.sessionManager.LoadAndSave)

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))
	router.Handler(http.MethodGet, "/", dynamicChain.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamicChain.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/snippet/create", dynamicChain.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", dynamicChain.ThenFunc(app.snippetCreatePost))

	// Create a middlware chain containing the standard middleware
	// which are to be used for every request our application receives.
	standardChain := alice.New(app.recoverPanic, app.logRequests, secureHeaders)

	return standardChain.Then(router)
}
