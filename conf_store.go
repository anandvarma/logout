package main

import (
	"log"

	bolt "go.etcd.io/bbolt"
)

// The ConfStore interface represents a persistent configuration store.
type ConfStore interface {
	AddOwner(owner string) error
	AddLog(key logKey, val logVal) error
	SetLogName(key logKey, name string) error
	RegisterWebPush(key logKey, secret string) error
}

// A confStore implementation that uses bolt db for persistence.
type boltConfStore struct {
	db *bolt.DB
}

func createBoltConfDB(path string) boltConfStore {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
		return boltConfStore{}
	}

	return boltConfStore{db: db}
}

func (bcs boltConfStore) AddOwner(owner string) error {
	return nil
}

func (bcs boltConfStore) AddLog(key logKey, val logVal) error {
	return nil
}

func (bcs boltConfStore) SetLogName(key logKey, name string) error {
	return nil
}

func (bcs boltConfStore) RegisterWebPush(key logKey, name string) error {
	return nil
}
