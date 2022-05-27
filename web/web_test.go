package web_test

import (
	"bytes"
	"distribkv/config"
	"distribkv/db"
	"distribkv/web"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func CreateShardDB(t *testing.T, idx int) *db.Database {
	t.Helper()

	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("db%d", idx))
	if err != nil {
		t.Fatalf("Couldn't create a temp db %d: %v", idx, err)
	}
	tmpFile.Close()
	name := tmpFile.Name()

	t.Cleanup(func() { os.Remove(name) })

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Couldn't create new database %q: %v", name, err)
	}
	t.Cleanup(func() { closeFunc() })

	return db
}

func CreateShardServer(t *testing.T, idx int, addrs map[int]string) (*db.Database, *web.Server) {
	t.Helper()

	db := CreateShardDB(t, idx)

	cfg := &config.Shards{
		Addrs:  addrs,
		Count:  len(addrs),
		CurIdx: idx,
	}

	s := web.NewServer(db, cfg)
	return db, s
}

func TestWebServer(t *testing.T) {
	var ts1GetHandler, ts1SetHandler func(w http.ResponseWriter, r *http.Request)
	var ts2GetHandler, ts2SetHandler func(w http.ResponseWriter, r *http.Request)

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/get") {
			ts1GetHandler(w, r)
		} else if strings.HasPrefix(r.RequestURI, "/set") {
			ts1SetHandler(w, r)
		}
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/get") {
			ts2GetHandler(w, r)
		} else if strings.HasPrefix(r.RequestURI, "/set") {
			ts2SetHandler(w, r)
		}
	}))
	defer ts2.Close()

	addrs := map[int]string{
		0: strings.TrimPrefix(ts1.URL, "http://"),
		1: strings.TrimPrefix(ts2.URL, "http://"),
	}

	db1, web1 := CreateShardServer(t, 0, addrs)
	db2, web2 := CreateShardServer(t, 1, addrs)

	keys := map[string]int{
		"Rus": 1,
		"USA": 0,
	}

	ts1GetHandler = web1.GetHandler
	ts1SetHandler = web1.SetHandler
	ts2GetHandler = web2.GetHandler
	ts2SetHandler = web2.SetHandler

	for key := range keys {
		_, err := http.Get(fmt.Sprintf(ts1.URL+"/set?key=%s&value=value-%s", key, key))
		if err != nil {
			t.Fatalf("Couldn't set the key %q: %v", key, err)
		}
	}

	for key := range keys {
		resp, err := http.Get(fmt.Sprintf(ts1.URL+"/get?key=%s", key))
		if err != nil {
			t.Fatalf("Get key %q error: %v", key, err)
		}
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Couldn't read contents of the key %q: %v", key, err)
		}

		want := []byte("value-" + key)
		if !bytes.Contains(contents, want) {
			t.Errorf("Unexpected contents of the key %q: got %q, want the result to contain %q", key, contents, want)
		}
		log.Printf("Contents of key %q: %s", key, err)
	}

	value1, err := db1.GetKey("USA")
	if err != nil {
		t.Fatalf("USA key error: %v", err)
	}

	want1 := "value-USA"
	if !bytes.Equal(value1, []byte(want1)) {
		t.Errorf("Unexpected value of key USA: got %q, want %q", value1, want1)
	}

	value2, err := db2.GetKey("Rus")
	if err != nil {
		t.Fatalf("Rus key error: %v", err)
	}

	want2 := "value-Rus"
	if !bytes.Equal(value2, []byte(want2)) {
		t.Errorf("Unexpected value of key USA: got %q, want %q", value2, want2)
	}
}
