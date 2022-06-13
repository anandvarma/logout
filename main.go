package main

import (
	"fmt"
)

func main() {
	logServer := NewLogServer(HOST + ":" + TCP_PORT)
	go logServer.Start()
	logServer.StartWeb()

	fmt.Println("Done")
}
