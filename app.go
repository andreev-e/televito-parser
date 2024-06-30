package main

import (
	"log"
	"os"
	Ssge "televito-parser/addsources/ssge"
	Dbmethods "televito-parser/dbmethods"
	Lrucache "televito-parser/lrucache"
	Routines "televito-parser/routines"
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

	for _, class := range []string{
		"MyAutoGe",
		"MyAutoGeRent",
		Ssge.Class,
		//Myhomege.Class
	} {
		go Routines.ReparseFirstPages(class)
		go Routines.ReparseAllPages(class)
	}

	for {
		log.Print(Dbmethods.GetDbStats())
		time.Sleep(1 * time.Minute)
	}
}
