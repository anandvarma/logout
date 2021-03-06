package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// The LogDb interface represents a persistent store of log buffers.
type LogDb interface {
	Commit(token int64, rb *RingBuf) bool
	GetLogBuf(token int64) []byte
}

// A mock no-op implementation of LogDb
type NoopLogDb struct {
}

func createNoopLogDb() *NoopLogDb {
	return &NoopLogDb{}
}

func (ldb *NoopLogDb) Commit(token int64, rb *RingBuf) bool {
	return true
}

func (ldb *NoopLogDb) GetLogBuf(token int64) []byte {
	return nil
}

// A LogDb implementation that uses the file system as a persistent store
// of logs.
type FSLogDb struct {
	path string
}

func createFSLogDb(path string) *FSLogDb {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Failed to open DB path: %s", path)
	}
	if !fi.IsDir() {
		log.Fatalf("DB path: %s is not a directory!", path)
	}

	return &FSLogDb{path: path}
}

func (ldb *FSLogDb) Commit(token int64, rb *RingBuf) bool {
	f, err := os.Create(ldb.GetFilePath(token))
	if err != nil {
		log.Printf("DB commit failed for %d", token)
		return false
	}
	defer f.Close()

	io.Copy(f, rb)
	f.Sync()
	return true
}

func (ldb *FSLogDb) GetFilePath(token int64) string {
	hexToken := fmt.Sprintf("%x", token)
	return filepath.Join(ldb.path, hexToken)
}

func (ldb *FSLogDb) GetLogBuf(token int64) []byte {
	buf, err := os.ReadFile(ldb.GetFilePath(token))
	if err != nil {
		log.Printf("DB read failed for %d", token)
		return nil
	}
	return buf
}
