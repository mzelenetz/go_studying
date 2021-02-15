package controllers

import (
	"fmt"
	"net/http"
	"../models"
	"../views"
	"../rand"
)

// Putting this here is for consistency later

//New Users is used to create a new users controller.
//This will panic if the templeates are not
//parsed correctly and should only be used during
//inital setup.
func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us: us,
	}
}

type Users struct {
	NewView *views.View
	LoginView *views.View
	us models.UserService

}

// New is used to render the form where a user can create a 
// new user account
//
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, nil)
}

type SignupForm struct {
	Name		string `schema:"name"`
	Email 		string `schema:"email"`
	Password    string `schema:"password"`
}

// Create processes the signup form when a user sumbits it
// Creates a new user account
//
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request){
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	user := models.User{
		Name: form.Name,
		Email: form.Email,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err := u.signIn(w, &user)
	if err != nil {
		// Check that this is correct
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

type LoginForm struct {
	Email string `schema:"email"`
	Password string `schema:"password"`
}

// Login is used to verify the provided email address and
// password and to log the user in if they are correct
//
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	user, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrPasswordIncorrect:
			fmt.Fprintln(w, "Invalid Password Provided")
		case models.ErrNotFound:
			fmt.Fprintln(w, "Invalid Email Address")
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}


	err = u.signIn(w, user)
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	// Make sure we have a remember token available
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name: "remember_token",
		Value: user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

// COokieTest is used to display cookies set on the current user
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request){
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)
}