package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type LogoutServer struct {
	pubsub  PubSub
	manager logManager
}

func NewLogoutServer() LogoutServer {
	// Seed rand to avoid collisions after a restart.
	rand.Seed(time.Now().UnixNano())

	ls := createFSLogDb(FS_LOG_DB_PATH)
	cs := createBoltConfDB(BOLT_CONF_DB_PATH)

	return LogoutServer{
		pubsub:  NewPubSub(),
		manager: newLogManager(ls, cs),
	}
}

func (ls *LogoutServer) Start() {
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

func (ls *LogoutServer) Shutdown() {
}
