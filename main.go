package main

import (
	"fmt"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ls := NewLogoutServer()
	ls.Run()
	fmt.Println("Shutting down...")
}
