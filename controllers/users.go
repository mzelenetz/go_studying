package controllers

import (
	"fmt"
	"net/http"
	"../views"
)

// Putting this here is for consistency later

//New Users is used to create a new users controller.
//This will panic if the templeates are not
//parsed correctly and should only be used during
//inital setup.
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
	}
}

type Users struct {
	NewView *views.View
}

// New is used to render the form where a user can create a 
// new user account
//
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, nil)
}

type SignupForm struct {
	Email 		string `json:"email"`
	Password    string `json:"password"`
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
	fmt.Fprintln(w, form)
}

