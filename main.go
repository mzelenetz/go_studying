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
	"./controllers"

	"github.com/gorilla/mux"
)
var (
	homeView *views.View
	contactView *views.View
)


func home(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

func contact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
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
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	usersC := controllers.NewUsers()

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/signup", usersC.New)
	r.HandleFunc("/faq", faq)
	http.ListenAndServe(":3000", r)
}

func must(err error){
	if err != nil{
		panic(err)
	}
}