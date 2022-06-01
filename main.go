package main

import (
	"fmt"
)

func main() {
	logServer := NewLogServer(HOST + ":" + PORT)
	go logServer.Start()
	logServer.StartWeb()

	fmt.Println("Done")
}
