package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
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

type PhotoData struct {
	UId  int
	Pos  int
	Size string
	Note string
	Name string
}

//TODO:the title shold be set in the photowall page
//you can just update it with the name you want, make
//it simple here
type PhotoAlbum struct {
	Title string
	Album []PhotoData
}

var database *sql.DB

var Store = sessions.NewCookieStore([]byte("hpb"))

func (p *PostData) WriteDb() error {
	stmt, e := database.Prepare("INSERT INTO postdata(username, content, created) VALUES(?,?,?)")
	if e != nil {
		log.Print(e)
		return e
	}
	res, e := stmt.Exec(p.UserName, p.Content, p.Created)
	if e != nil {
		log.Print(e)
		return e
	}
	lastId, e := res.LastInsertId()
	if e != nil {
		log.Print(e)
		return e
	}
	rowCnt, e := res.RowsAffected()
	if e != nil {
		log.Print(e)
		return e
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
	return e
}
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//the := can't be used for er
	var er error
	database, er =
		sql.Open("sqlite3", "./postData.db")
	if er != nil {
		log.Fatal(er)
	}

	sql_PostTable := `
    CREATE TABLE IF NOT EXISTS postdata(
        uid INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(64) NULL,
        content VARCHAR(3000) NULL,
        created VARCHAR(64) NULL
    );
    `
	sql_PhotoTable := `
    CREATE TABLE IF NOT EXISTS photodata(
        uid INTEGER PRIMARY KEY AUTOINCREMENT,
        pos INTEGER NULL,
        size VARCHAR(10) NULL,
        note VARCHAR(64) NULL,
        name VARCHAR(3000) NULL
    );
    `
	_, er = database.Exec(sql_PostTable)
	if er != nil {
		log.Fatal(er)
	}
	_, er = database.Exec(sql_PhotoTable)
	if er != nil {
		log.Fatal(er)
	}
}
func delPostdata(sqlstr string) error {
	//stmt, e = db.Prepare("delete from userinfo where uid=?")
	stmt, e := database.Prepare(sqlstr)
	if e != nil {
		log.Print(e)
		return e
	}

	res, e := stmt.Exec()
	if e != nil {
		log.Print(e)
		return e
	}

	affect, e := res.RowsAffected()
	if e != nil {
		log.Print(e)
		return e
	}

	fmt.Println(affect)

	return e
}

func findPostdata(sqlstr string) ([]PostData, error) {
	var s []PostData
	rows, e := database.Query(sqlstr)
	if e != nil {
		log.Fatal(e)
	}
	var uid int
	var username string
	var content string
	var created string

	for rows.Next() {
		e = rows.Scan(&uid, &username, &content, &created)
		if e != nil {
			log.Fatal(e)
		}
		p := PostData{UId: uid, UserName: username, Content: content, Created: created}
		s = append(s, p)
	}

	rows.Close()

	return s, e
}
func showPostBoard(pattern string, w http.ResponseWriter) error {
	sqlStr := "SELECT " + pattern + " FROM postdata"
	cc, e := findPostdata(sqlStr)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return e
	}
	t, e := template.ParseFiles("./templates/board.html")
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return e
	}
	e = t.Execute(w, PostContext{Context: cc})
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return e
	}
	return e
}

//IsAct will check if the user has an active session and return True
func IsAct(r *http.Request) bool {
	session, _ := Store.Get(r, "session")
	if session.Values["act"] == "true" {
		return true
	}
	return false
}
func tokenCreate() string {
	ct := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(ct, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))
	// fmt.Println("token created :", token)
	return token
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	var e error
	if r.Method == "GET" {
		session, _ := Store.Get(r, "session")

		if session.Values["act"] != "true" {
			session.Values["act"] = "true"
			e = session.Save(r, w)
			if e != nil {
				log.Fatal(e)
			}
			fmt.Println("write the session data")
		}

		e = showPostBoard("*", w)
		if e != nil {
			log.Print(e)
		}
	} else if r.Method == "POST" {
		session, e := Store.Get(r, "session")
		if e != nil {
			log.Fatal(e)
		}
		if session.Values["act"] == "true" {
			e = r.ParseForm()
			if e != nil {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}

			t := time.Now().Format("2006-01-02 15:04:05")
			n := strings.Split(r.RemoteAddr, ":")[0] + "-" + strings.TrimLeft(strings.Fields(r.UserAgent())[1], "(")
			uname := strings.TrimRight(n, ";")
			// fmt.Println("uanme =", uname)
			p := PostData{UserName: uname, Content: r.Form["body"][0], Created: t}
			e = p.WriteDb()
			if e != nil {
				log.Print(e)
			}

			e = showPostBoard("*", w)
			if e != nil {
				log.Print(e)
			}
			// session.Values["username"] = username

		} else {
			fmt.Println("the session is not active go to root")
			http.Redirect(w, r, "/", 302)
		}
	} else {
		http.Error(w, "Unknown HTTP Action", http.StatusInternalServerError)
		return

	}
}

func AddPhotoHandler(w http.ResponseWriter, r *http.Request) {
	var e error
	fmt.Println("the r.methond is", r.Method)
	if r.Method == "GET" {
		t, e := template.ParseFiles("./templates/addphoto.html")
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		e = t.Execute(w, nil)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		e = r.ParseForm()
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		file, handler, e := r.FormFile("uploadfile")
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		if handler != nil {
			r.ParseMultipartForm(32 << 20) //defined maximum size of file
			defer file.Close()
			rand := md5.New()
			io.WriteString(rand, strconv.FormatInt(time.Now().Unix(), 10))
			io.WriteString(rand, handler.Filename)
			tkn := fmt.Sprintf("%x", rand.Sum(nil))
			f, e := os.OpenFile("./files/images/"+tkn, os.O_WRONLY|os.O_CREATE, 0666)
			if e != nil {
				log.Println(e)
				return
			}
			defer f.Close()
			io.Copy(f, file)
			fmt.Println("upload a file done!")
			//filelink := "<br> <a href=./files/" + handler.Filename + ">" + handler.Filename + "</a>"
		}
		//n := strings.Split(r.RemoteAddr, ":")[0] + "-" + strings.TrimLeft(strings.Fields(r.UserAgent())[1], "(")
		//uname := strings.TrimRight(n, ";")
		//p := PostData{UserName: uname, Content: r.Form["body"][0], Created: t}
		//p.WriteDb()
		http.Redirect(w, r, "/photowall", 302)
	} else {
		log.Print("Unknown request")
		http.Redirect(w, r, "/photowall", 302)
	}
}
func PhotoWallHandler(w http.ResponseWriter, r *http.Request) {
	t, e := template.ParseFiles("./templates/photoWall.html")
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	e = t.Execute(w, nil)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

}
func AddPostHandler(w http.ResponseWriter, r *http.Request) {
	var e error
	if r.Method == "GET" {
		t, e := template.ParseFiles("./templates/addpost.html")
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		token := tokenCreate()
		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "csrftoken", Value: token, Expires: expiration}
		http.SetCookie(w, &cookie)
		e = t.Execute(w, token)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		e = r.ParseForm()
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		//
		formToken := template.HTMLEscapeString(r.Form.Get("CSRFToken"))
		cookie, e := r.Cookie("csrftoken")
		if e != nil {
			log.Print(e)
			return
		}
		if formToken == cookie.Value {

			t := time.Now().Format("2006-01-02 15:04:05")
			n := strings.Split(r.RemoteAddr, ":")[0] + "-" + strings.TrimLeft(strings.Fields(r.UserAgent())[1], "(")
			uname := strings.TrimRight(n, ";")
			// fmt.Println("uanme =", uname)
			p := PostData{UserName: uname, Content: r.Form["body"][0], Created: t}
			p.WriteDb()
		} else {
			log.Print("form token mismatch")
		}
		http.Redirect(w, r, "/", 302)
	} else {
		log.Print("Unknown request")
		http.Redirect(w, r, "/", 302)
	}

}
func main() {
	defer database.Close()
	http.HandleFunc("/addpost/", AddPostHandler)
	http.HandleFunc("/addphoto/", AddPhotoHandler)
	http.HandleFunc("/photowall/", PhotoWallHandler)
	http.HandleFunc("/", rootHandler)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates"))))
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files"))))
	log.Print("Running the server on port 8091.")
	log.Fatal(http.ListenAndServe(":8091", nil))
}
