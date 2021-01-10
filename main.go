//Studying Go development by following Jon Calhoun's course. 
package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if r.URL.Path == "/" {
		fmt.Fprint(w, "<h1>This is my Homepage<h1>")
		fmt.Fprint(w, "<a href=\"/contact\">Contact")

	} else if r.URL.Path == "/contact" {
		fmt.Fprint(w, "Contact: <a href=\"mailto: mzelenetz@gmail.com\">mzelenetz@gmail.com</a>.")
	}
}

func main(){
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":3000", nil)
}