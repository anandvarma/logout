package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// The logStore interface represents a persistent store of log buffers.
type logStore interface {
	CommitLog(key logKey, val logVal) error
	GetClosedLogReader(key logKey) (io.ReadCloser, error)
	DeleteLog(key logKey) error
}

// A LogStore implementation that uses the file system as a persistent store
// of logs.
type FSLogDb struct {
	path string
}

func createFSLogDb(path string) *FSLogDb {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Failed to open DB path: %s", path)
		return nil
	}
	if !fi.IsDir() {
		log.Fatalf("DB path: %s is not a directory!", path)
		return nil
	}

	return &FSLogDb{path: path}
}

func (ldb *FSLogDb) CommitLog(key logKey, val logVal) error {
	f, err := os.Create(ldb.getFilePath(key))
	if err != nil {
		log.Printf("DB commit failed for %v", key)
		return ErrKeyExists
	}
	defer f.Close()

	io.Copy(f, &val.rb)
	f.Sync()
	return nil
}

func (ldb *FSLogDb) GetClosedLogReader(key logKey) (io.ReadCloser, error) {
	f, err := os.Open(ldb.getFilePath(key))
	if err != nil {
		return nil, ErrLogNotFound
	}
	return f, nil
}

func (ldb *FSLogDb) DeleteLog(key logKey) error {
	err := os.Remove(ldb.getFilePath(key))
	if err != nil {
		return ErrLogNotFound
	}
	return nil
}

func (ldb *FSLogDb) getFilePath(key logKey) string {
	hexToken := fmt.Sprintf("%x", key.token)
	return filepath.Join(ldb.path, hexToken)
}
