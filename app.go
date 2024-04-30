package main

import (
	"log"
	"os"
	Myhomege "televito-parser/addsources/myhomege"
	Ssge "televito-parser/addsources/ssge"
	Dbmethods "televito-parser/dbmethods"
	Lrucache "televito-parser/lrucache"
	"time"
)

func init() {
	Dbmethods.InitDB()
	Lrucache.CachedLocations = Lrucache.Constructor(1000)
}

func main() {
	logFile, err := os.Create("1.log")
	if err != nil {
		log.Println("Error creating log file:", err)
		return
	}

	log.SetOutput(logFile)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
	}()
	defer logFile.Close()
	defer Dbmethods.CloseDB()

	//for _, class := range []string{"MyAutoGe", "MyAutoGeRent", Ssge.Class, Myhomege.Class} {
	//	go reparseFirstPages(class)
	//	go reparseAllPages(Myhomege.Class)
	//}

	go reparseFirstPages("MyAutoGe")
	go reparseFirstPages("MyAutoGeRent")
	go reparseFirstPages(Ssge.Class)
	go reparseFirstPages(Myhomege.Class)

	go reparseAllPages("MyAutoGe")
	go reparseAllPages("MyAutoGeRent")
	go reparseAllPages(Ssge.Class)
	go reparseAllPages(Myhomege.Class)

	for {
		log.Print(Dbmethods.GetDbStats())
		time.Sleep(1 * time.Minute)
	}
}
