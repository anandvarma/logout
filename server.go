package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type LogoutServer struct {
	pubsub PubSub
	store  LogStore
}

func NewLogoutServer() LogoutServer {
	// Seed rand to avoid collisions after a restart.
	rand.Seed(time.Now().UnixNano())

	return LogoutServer{
		pubsub: NewPubSub(),
		store:  NewLogStore(),
	}
}

func (ls *LogoutServer) Run() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Start TCP handler.
	go func() {
		defer wg.Done()
		ls.startTcpServer(HOST + ":" + TCP_PORT)
	}()

	// Start web handler.
	go func() {
		defer wg.Done()
		ls.startWebServer(HOST + ":" + WEB_PORT)
	}()

	wg.Wait()
	log.Print("LogoutServer Run(): WaitGroup done.")
}
