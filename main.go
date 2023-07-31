package main

import (
	"fmt"
)

func main() {
	ls := NewLogoutServer()
	ls.Run()
	fmt.Println("Shutting down...")
}
