package main

import (
	"time"
)

func main() {
	initDB()
	defer CloseDB()

	go reparseFirstPages()
	go reparseAllPages()
	for {
		time.Sleep(10 * time.Second)
	}
}
