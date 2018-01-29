package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAddPostHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(AddPostHandler))
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Errorf("Error occured while constructing request: %s", err)
	}

	w := httptest.NewRecorder()
	AddPostHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Actual status: (%d); Expected status:(%d)", w.Code, http.StatusOK)
	}
}

func TestRootHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(rootHandler))
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Errorf("Error occured while constructing request: %s", err)
	}

	w := httptest.NewRecorder()
	rootHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Actual status: (%d); Expected status:(%d)", w.Code, http.StatusOK)
	}
}

//TestInitTableDatabase is to make sure the data table structure is right
func TestInitTableDataBase(t *testing.T) {
	sqlStr := "select name from sqlite_master where type='table' order by name;"
	expect := "photodata-postdata-sqlite_sequence-"
	var actual string
	rows, err := database.Query(sqlStr)
	if err != nil {
		t.Fatalf("Error occured while checking DB: %s", err)
	}
	var tabelName string

	for rows.Next() {
		err = rows.Scan(&tabelName)
		if err != nil {
			t.Errorf("Error occured while checking DB: %s", err)
		}
		actual += tabelName
		actual += "-"
	}

	rows.Close()
	if expect != actual {
		t.Errorf("Error occured while checking DB tables")
	}
}

//TestPostWRD is to test the db data can be read/write/del
func TestPostWRD(t *testing.T) {
	day := time.Now().Format("2006-01-02 15:04:05")
	delstr := "delete from postdata where username='testpostwrd';"
	p := PostData{UserName: "testpostwrd", Content: "testpostwrd-content", Created: day}
	err := p.WriteDb()
	if err != nil {
		t.Errorf("Error writing the database : %s", err)
	}

	sqlStr := "select * from postdata where username='testpostwrd';"
	s, err := findPostdata(sqlStr)
	if err != nil {
		t.Errorf("Error retriving the data")
	}
	if len(s) != 1 {
		//need to clean up for some underlying issue
		for i := 0; i < len(s)-1; i++ {
			delPostdata(delstr)
		}
		t.Error("Error the data creating size is not 1")
	} else if s[0].Content != "testpostwrd-content" {
		t.Errorf("Error the database mismatch %s", s[0].Content)
	}

	err = delPostdata(delstr)
	if err != nil {
		t.Errorf("Error deleting the database : %s", err)
	}

}
