package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/vladComan0/go-snippets/internal/models"
	"github.com/vladComan0/go-snippets/internal/validator"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type accountPasswordUpdateForm struct {
	CurrentPassword         string `form:"currentPassword"`
	NewPassword             string `form:"newPassword"`
	NewPasswordConfirmation string `form:"newPasswordConfirmation"`
	validator.Validator     `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
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
	data := app.newTemplateData(r)
	data.Snippet = snippet
	app.render(w, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = &snippetCreateForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validating untrusted user input
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank.")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long.")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank.")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 30, 365), "expires", "This field must equal 1, 7, 30 or 365.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	const PASSWORD_LENGTH = 8
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank.")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid e-mail address.")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank.")
	form.CheckField(validator.MinChars(form.Password, PASSWORD_LENGTH), "password", "This field must be at least 8 characters long.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	if err := app.users.Insert(form.Name, form.Email, form.Password); err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateEmail):
			form.AddFieldError("email", "E-mail address already used.")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		default:
			app.serverError(w, err)
		}
		return
	}

	// Otherwise add a confirmation flash message to the session and redirect to the login page
	app.sessionManager.Put(r.Context(), "flash", "You have signed up successfully. You can login now.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid e-mail address.")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// non-field error message and re-display the login page.
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidCredentials):
			form.AddNonFieldError("Email or password is incorrect.")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		default:
			app.serverError(w, err)
		}
		return
	}

	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the user (e.g. login
	// and logout operations).
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Check for the existence of the redirectPathAfterLogin value in the session.
	// If the value exists, use it to redirect the user to that URL. Then delete
	// the value from the session.
	redirectPathAfterLogin := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if redirectPathAfterLogin != "" {
		http.Redirect(w, r, redirectPathAfterLogin, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, err)
		return
	}

	// Remove the authenticatedUserID from session data so that the user
	// is logged out
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully.")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.tmpl.html", data)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	user, err := app.users.Get(userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
	}

	data := app.newTemplateData(r)
	data.User = user

	app.render(w, http.StatusOK, "account.tmpl.html", data)
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}

	app.render(w, http.StatusOK, "password.tmpl.html", data)
}

func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form accountPasswordUpdateForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	const PASSWORD_LENGTH = 8

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank.")

	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank.")
	form.CheckField(validator.MinChars(form.NewPassword, PASSWORD_LENGTH), "newPassword", "This field must be at least 8 characters long.")

	form.CheckField(validator.NotBlank(form.NewPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank.")
	form.CheckField(validator.Compare(form.NewPassword, form.NewPasswordConfirmation), "newPasswordConfirmation", "Passwords do not match.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "password.tmpl.html", data)
		return
	}

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if err := app.users.UpdatePassword(userID, form.CurrentPassword, form.NewPassword); err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidCredentials):
			form.AddFieldError("currentPassword", "Current password is incorrect.")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "password.tmpl.html", data)
		case errors.Is(err, models.ErrSamePassword):
			form.AddFieldError("newPassword", "New password cannot be the same as the current password.")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "password.tmpl.html", data)
		default:
			app.serverError(w, err)
		}
		return
	}

	// Add a confirmation flash message to the session and redirect to the login page
	app.sessionManager.Put(r.Context(), "flash", "Your password has been updated successfully!")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}
