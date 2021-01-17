package controllers

import (
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
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

type Users struct {
	NewView *views.View
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, nil)
}