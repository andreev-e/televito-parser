package main

import (
	Dbmethods "televito-parser/dbmethods"
	"time"
)

func main() {
	Dbmethods.InitDB()
	defer Dbmethods.CloseDB()

	go reparseFirstPages("MyAutoGe")
	go reparseFirstPages("MyAutoGeRent")
	go reparseAllPages("MyAutoGe")
	go reparseAllPages("MyAutoGeRent")
	for {
		time.Sleep(10 * time.Second)
	}
}
