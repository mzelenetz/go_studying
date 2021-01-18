//Studying Go development by following Jon Calhoun's course. 
package main

// Routers Jon Likes: "julienschmidt httprouter"
// Features:
	// you can get the parameter as a slice: 

// gorilla/mux
import (
	"net/http"

	"./controllers"

	"github.com/gorilla/mux"
)

func main(){
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers()

	r := mux.NewRouter()
	r.Handle("/", staticC.HomeView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	http.ListenAndServe(":3000", r)
}

func must(err error){
	if err != nil{
		panic(err)
	}
}