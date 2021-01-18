package controllers

import "../views"

func NewStatic() *Static{
	return &Static{
		HomeView: views.NewView("bootstrap", "views/static/home.gohtml"),
		ContactView: views.NewView("bootstrap", "views/static/contact.gohtml"),
	}
}

type Static struct {
	HomeView *views.View 
	ContactView *views.View 
}