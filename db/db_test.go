package db_test

import (
	"bytes"
	"distribkv/db"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetSet(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "kvdb0")
	if err != nil {
		t.Fatalf("Couldn't create temp file: %v", err)
	}

	name := f.Name()
	f.Close()
	defer os.Remove(name)

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Couldn't create a new database: %v", err)
	}
	defer closeFunc()

	if err := db.SetKey("test", []byte("Great")); err != nil {
		t.Fatalf("Couldn't write key: %v", err)
	}

	value, err := db.GetKey("test")
	if err != nil {
		t.Fatalf(`Couldn't get the key "test": %v`, err)
	}

	if !bytes.Equal(value, []byte("Great")) {
		t.Errorf(`Unexpected value for key "test": got: %q, want %q`, value, "Greats")
	}
}
