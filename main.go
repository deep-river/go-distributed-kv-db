package main

import (
	"distribkv/db"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	dbLocation = flag.String("db-location", "my.db", "The bolt db databasse filepath")
	httpAddr   = flag.String("http-addr", "127.0.0.1:8080", "http host and port")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatal("Warning: Must provide db-location")
	}
}

func main() {
	parseFlags()

	db, close, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("NewDatabase(%q): %v", *dbLocation, err)
	}
	defer close()

	fmt.Println("distribkv server start")

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		key := r.Form.Get("key")
		value, err := db.GetKey(key)
		fmt.Fprintf(w, "called GET, Value = %q, error = %v", value, err)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		key := r.Form.Get("key")
		value := r.Form.Get(("value"))
		err := db.SetKey(key, []byte(value))
		fmt.Fprintf(w, "called SET, error = %v", err)
	})

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
