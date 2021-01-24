//Studying Go development by following Jon Calhoun's course. 
package main

// Routers Jon Likes: "julienschmidt httprouter"
// Features:
	// you can get the parameter as a slice: 

// gorilla/mux
import (
	"net/http"
	"./models"
	"./controllers"
	"fmt"

	"github.com/gorilla/mux"
)

// Will remove the passwords later
const (
	host = "localhost"
	port = 5432
	user = "postgres"
	password = "alexnoah"
	dbname = "databot_dev"
)

func main(){
	// Connect to UserService Model
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

	us, err := models.NewUserService(psqlInfo)
	must(err)
	defer us.Close()
	us.AutoMigrate()
 
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(us)

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