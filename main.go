package main

import (
	"distribkv/config"
	"distribkv/db"
	"distribkv/web"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
)

// Currently using flag - commandline parameters for configuration
// TODO: Use config files instead
var (
	dbLocation = flag.String("db-location", "my.db", "The bolt db databasse filepath")
	httpAddr   = flag.String("http-addr", "127.0.0.1:8080", "http host and port")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
	shard      = flag.String("shard", "Moscow", "Name of the shard for storage")
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

	var c config.Config
	if _, err := toml.DecodeFile(*configFile, &c); err != nil {
		log.Fatalf("toml.DecodeFile(%q): %v", *configFile, err)
	}

	log.Printf("%#v", &c)

	var shardCount int
	var shardIdx int = -1
	for _, s := range c.Shards {
		if s.Idx+1 > shardCount {
			shardCount = s.Idx + 1
		}
		if s.Name == *shard {
			shardIdx = s.Idx
		}
	}

	if shardIdx < 0 {
		log.Fatalf("Shard %q was not found", *shard)
	}

	log.Printf("Shard count is %d, curreent shard: %d", shardCount, shardIdx)

	db, close, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("NewDatabase(%q): %v", *dbLocation, err)
	}
	defer close()

	fmt.Println("distribkv server start")

	srv := web.NewServer(db, shardIdx, shardCount)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
