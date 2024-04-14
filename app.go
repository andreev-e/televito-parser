package main

import (
	"time"
)

func main() {
	go reparseFirstPages()
	for {
		time.Sleep(10 * time.Second)
	}
}
