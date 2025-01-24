package main

import "runtime"

func main() {
	runTimeOs := runtime.GOOS
	server := NewServer("127.0.0.1", 8888, runTimeOs)
	server.Start()
}
