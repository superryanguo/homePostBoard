package main

import (
	"database/sql"
	"fmt"
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
	fmt.Fprintf(w, "RequestURI: %s\n", r.RequestURI)
	fmt.Fprintf(w, "RequestBody: %s\n", r.Body)
}
func main() {
	http.Handle("/static/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/", rootHandler)
	log.Print("Running the server on port 8091.")
	log.Fatal(http.ListenAndServe(":8091", nil))
}
