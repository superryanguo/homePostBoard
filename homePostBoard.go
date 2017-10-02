package main

import (
	"fmt"
	"log"
	"net/http"
)

type PostData struct {
	userId string
	post   string
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "RequestURI: %s\n", r.RequestURI)
}
func main() {
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(":8091", nil))
}
