package homepostboard

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
