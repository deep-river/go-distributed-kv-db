package web

import (
	"distribkv/db"
	"fmt"
	"hash/fnv"
	"net/http"
)

type Server struct {
	db         *db.Database
	shardIdx   int
	shardCount int
}

func NewServer(db *db.Database, shardIdx, shardCount int) *Server {
	return &Server{
		db:         db,
		shardIdx:   shardIdx,
		shardCount: shardCount,
	}
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	value, err := s.db.GetKey(key)
	fmt.Fprintf(w, "called GET, Value = %q, error = %v", value, err)
}

func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	value := r.Form.Get(("value"))

	h := fnv.New64()
	h.Write([]byte(key))
	shardIdx := int(h.Sum64() % uint64(s.shardCount))

	err := s.db.SetKey(key, []byte(value))
	fmt.Fprintf(w, "called SET, error = %v, hash = %d, shardIdx = %d", err, h.Sum64(), shardIdx)
}
