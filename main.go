package main

import (
	"distribkv/config"
	"distribkv/db"
	"distribkv/web"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	dbLocation = flag.String("db-location", "my.db", "The bolt db databasse filepath")
	httpAddr   = flag.String("http-addr", "127.0.0.1:8080", "http host and port")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
	shard      = flag.String("shard", "Moscow", "Name of the shard for storage")
	replica    = flag.Bool("replica", false, "Whether or not run as a read-only replica")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatal("Warning: Must provide db-location")
	}

	if *shard == "" {
		log.Fatal("Warning: Must provide shard")
	}
}

func main() {
	parseFlags()

	c, err := config.ParseFile(*configFile)
	if err != nil {
		log.Fatalf("Error parsing config %q: %v", *configFile, err)
	}

	shards, err := config.ParseShards(c.Shards, *shard)
	if err != nil {
		log.Fatalf("Error parsing shards config: %v", err)
	}

	log.Printf("Shard count is %d, curreent shard: %d", shards.Count, shards.CurIdx)

	db, close, err := db.NewDatabase(*dbLocation, *replica)
	if err != nil {
		log.Fatalf("NewDatabase(%q): %v", *dbLocation, err)
	}
	defer close()

	fmt.Println("distribkv server start")

	srv := web.NewServer(db, shards)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)
	http.HandleFunc("/purge", srv.DeleteExtraKeysHandler)
	http.HandleFunc("/next-replication-key", srv.GetNextKeyForReplication)
	http.HandleFunc("/delete-replication-key", srv.DeleteReplicationKey)

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
