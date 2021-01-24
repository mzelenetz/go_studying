package controllers

import (
	"fmt"
	"net/http"
	"../models"
	"../views"
)

// Putting this here is for consistency later

//New Users is used to create a new users controller.
//This will panic if the templeates are not
//parsed correctly and should only be used during
//inital setup.
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
		us: us,
	}
}

type Users struct {
	NewView *views.View
	us *models.UserService
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
	}

	if err := u.us.Create(&user); err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintln(w, form)
}

