package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/vladComan0/letsgo/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := &templateData{
		Snippets: snippets,
	}
	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	snippet, err := app.snippets.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNoRecord):
			app.clientError(w, http.StatusNotFound)
		default:
			app.serverError(w, err)
		}
		return
	}
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/view.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		fmt.Println("heyo")
		app.serverError(w, err)
		return
	}
	data := &templateData{
		Snippet: snippet,
	}
	if err = ts.ExecuteTemplate(w, "base", data); err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
