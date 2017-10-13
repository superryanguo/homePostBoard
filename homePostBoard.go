package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//TODO:UID is for IP and Name is for requestBody
type PostData struct {
	UId      int
	UserName string
	Content  string
	Created  string
}
type PostContext struct {
	Context []PostData
}

var database *sql.DB
var err error

func (p *PostData) WriteDb() {
	stmt, err := database.Prepare("INSERT INTO postdata(username, content, created) VALUES(?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(p.UserName, p.Content, p.Created)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
}
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
func showPostBoard(pattern string, w http.ResponseWriter) (err error) {
	sqlStr := "SELECT " + pattern + " FROM postdata"
	rows, err := database.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}
	var c PostContext
	var uid int
	var username string
	var content string
	var created string

	for rows.Next() {
		err = rows.Scan(&uid, &username, &content, &created)
		if err != nil {
			log.Fatal(err)
		}
		p := PostData{UId: uid, UserName: username, Content: content, Created: created}
		c.Context = append(c.Context, p)
	}

	rows.Close()

	t, err := template.ParseFiles("./templates/board.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "RequestURI: %s\n", r.RequestURI)
	// fmt.Fprintf(w, "RequestRemoteAddr: %s\n", r.RemoteAddr)
	// fmt.Fprintf(w, "RequestHeader: %s\n", r.Header)
	if r.Method == "GET" {
		err = showPostBoard("*", w)
		if err != nil {
			log.Fatal(err)
		}
	} else if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t := time.Now().Format("2006-01-02 15:04:05")
		p := PostData{UserName: "guo", Content: r.Form["body"][0], Created: t}
		p.WriteDb()

		err = showPostBoard("*", w)
		if err != nil {
			log.Fatal(err)
		}
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
