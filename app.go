package main

import (
	"fmt"
	"log"
	"os"
	Dbmethods "televito-parser/dbmethods"
	"time"
)

func main() {
	logFile, err := os.Create("logfile.log")
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
	}()

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
