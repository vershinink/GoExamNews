package server

import (
	"GoNews/internal/config"
	"GoNews/internal/storage"
	"GoNews/internal/storage/memdb"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var cfg = config.Config{
	HTTPServer: config.HTTPServer{
		Address:      "localhost:80",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 30,
	},
}
var posts = []storage.Post{
	{Title: "First post"},
	{Title: "Second post"},
}
var st = memdb.New()

func TestIndex(t *testing.T) {
	srv := New(&cfg)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.API(st)

	srv.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("incorrect code: got = %v, want = %v", rr.Code, http.StatusOK)
	}
}

func TestPosts(t *testing.T) {

	ch := make(chan storage.Post, len(posts))
	for _, p := range posts {
		ch <- p
	}
	close(ch)
	num, err := st.AddPosts(context.Background(), ch)
	if err != nil {
		t.Fatalf("error on adding posts to storage")
	}
	if num != len(posts) {
		t.Fatalf("error on adding posts to storage")
	}

	str := fmt.Sprintf("/news/%d", num)

	srv := New(&cfg)
	req := httptest.NewRequest(http.MethodGet, str, nil)
	rr := httptest.NewRecorder()
	srv.API(st)

	srv.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("incorrect code: got = %v, want = %v", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Errorf("cannot read the response body")
	}

	var got []Response
	err = json.Unmarshal(b, &got)
	if err != nil {
		t.Errorf("cannot decode posts")
	}
	if len(got) != num {
		t.Errorf("incorrect posts number: got = %d, want = %d", len(got), num)
	}
}
