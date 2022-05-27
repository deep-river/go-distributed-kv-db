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

func setKey(t *testing.T, d *db.Database, key, value string) {
	t.Helper()

	if err := d.SetKey(key, []byte(value)); err != nil {
		t.Fatalf("SetKey(%q, %q) failed: %v", key, value, err)
	}
}

func getKey(t *testing.T, d *db.Database, key string) string {
	t.Helper()

	value, err := d.GetKey(key)
	if err != nil {
		t.Fatalf("GetKey(%q) failed: %v", key, err)
	}
	return string(value)
}

func TestDeleteExtraKeys(t *testing.T) {
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

	setKey(t, db, "test", "Great")
	setKey(t, db, "data", "stays")

	if err := db.DeleteExtraKeys(func(name string) bool { return name == "test" }); err != nil {
		t.Fatalf("Couldn't delete exra keys: %v", err)
	}

	if value := getKey(t, db, "test"); value != "" {
		t.Errorf(`Unxpected value for key "test": got %q, want %q`, value, "")
	}
}
