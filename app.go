package main

import (
	"time"
)

func main() {
	go reparseFirstPages()
	go reparseAllPages()
	for {
		time.Sleep(10 * time.Second)
	}
}
