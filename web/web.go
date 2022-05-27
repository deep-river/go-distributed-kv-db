package web

import (
	"distribkv/config"
	"distribkv/db"
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	db     *db.Database
	shards *config.Shards
}

func NewServer(db *db.Database, s *config.Shards) *Server {
	return &Server{
		db:     db,
		shards: s,
	}
}

func (s *Server) redirect(shard int, w http.ResponseWriter, r *http.Request) {
	url := "http://" + s.shards.Addrs[shard] + r.RequestURI
	fmt.Fprintf(w, "redirecting from shard %d to shard %d (%q)\n", s.shards.CurIdx, shard, url)

	resp, err := http.Get(url)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error redirecting the request: %v", err)
		return
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	shard := s.shards.Index(key)
	value, err := s.db.GetKey(key)

	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}
	fmt.Fprintf(w, "called GET, Shard = %d, current shard = %d, addr = %q, Value = %q, error = %v", shard, s.shards.CurIdx, s.shards.Addrs[shard], value, err)
}

func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	value := r.Form.Get(("value"))

	shard := s.shards.Index(key)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	err := s.db.SetKey(key, []byte(value))
	fmt.Fprintf(w, "called SET, error = %v, shardIdx = %d, current shard = %d", err, shard, s.shards.CurIdx)
}

func (s *Server) DeleteExtraKeysHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Error = %v", s.db.DeleteExtraKeys(func(key string) bool {
		return s.shards.Index(key) != s.shards.CurIdx
	}))
}
