package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	// Configure the logger.
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ls := NewLogoutServer()

	// Setup interrupt handler to gracefully shutdown the server.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("captured %v, shutting down ...", sig)
			ls.Shutdown()
			os.Exit(1)
		}
	}()

	ls.Start()
}
