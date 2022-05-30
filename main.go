package main

func main() {
	logServer := NewLogServer(HOST + ":" + PORT)
	logServer.Start()
}
