//Studying Go development by following Jon Calhoun's course. 
package main

// Routers Jon Likes: "julienschmidt httprouter"
// Features:
	// you can get the parameter as a slice: 

// gorilla/mux
import (
	"fmt"
	"net/http"

	"./views"

	"github.com/gorilla/mux"
)
var (
	homeView *views.View
	contactView *views.View
)


func home(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	err := homeView.Template.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func contact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	err := contactView.Template.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func faq(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h2>Q: What is Your Name?</h2>")
	fmt.Fprint(w, "<h4>Q: Michael Zelenetz</h4>")
}

func notFound(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h2>Womp! Womp!</h2>")
	fmt.Fprint(w, "<h4>Page not found :(</h4>")
}

func main(){
	homeView = views.NewView("views/home.gohtml")
	contactView = views.NewView("views/contact.gohtml")

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/faq", faq)
	http.ListenAndServe(":3000", r)
}