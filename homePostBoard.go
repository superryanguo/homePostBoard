package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

//TODO:UID is for IP and Name is for requestBody
type PostData struct {
	UId      int
	UserName string
	Content  string
	Created  string
}

var database *sql.DB
var err error

func init() {
	database, err =
		sql.Open("sqlite3", "./postData.db")
	if err != nil {
		log.Fatal(err)
	}

	sql_table := `
    CREATE TABLE IF NOT EXISTS postdata(
        uid INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(64) NULL,
        content VARCHAR(3000) NULL,
        created DATE NULL
    );
    `
	_, err := database.Exec(sql_table)
	if err != nil {
		log.Fatal(err)
	}
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "RequestURI: %s\n", r.RequestURI)
	// fmt.Fprintf(w, "RequestRemoteAddr: %s\n", r.RemoteAddr)
	// fmt.Fprintf(w, "RequestHeader: %s\n", r.Header)
	if r.Method == "GET" {
		p := PostData{UId: 1, UserName: "Ryan", Content: "The First post from Ryan", Created: "20171010"}
		t, err := template.ParseFiles("./templates/board.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "RequestBody: %s\n", r.Form["body"])

	} else {
		http.Error(w, "Unknown HTTP Action", http.StatusInternalServerError)
		return

	}
}

func addPostHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/addpost.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func main() {
	defer database.Close()
	http.HandleFunc("/addpost/", addPostHandler)
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.FileServer(http.Dir("public")))
	log.Print("Running the server on port 8091.")
	log.Fatal(http.ListenAndServe(":8091", nil))
}
