package main

import (
	Dbmethods "televito-parser/dbmethods"
	"time"
)

func main() {
	Dbmethods.InitDB()
	defer Dbmethods.CloseDB()

	go reparseFirstPages()
	go reparseAllPages("MyAutoGe")
	for {
		time.Sleep(10 * time.Second)
	}
}
