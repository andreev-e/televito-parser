package main

import (
	"log"
	"strconv"
	"televito-parser/addsources/myautoge"
	Myhomege "televito-parser/addsources/myhomege"
	Ssge "televito-parser/addsources/ssge"
	Dbmethods "televito-parser/dbmethods"
	"time"
)

func reparseFirstPages(class string) {
	defer func() {
		log.Println("reparseFirstPages ended " + class)
	}()

	for {
		err := error(nil)
		switch class {
		case "MyAutoGe", "MyAutoGeRent":
			_, err = Myautoge.ParsePage(1, class)
		case Ssge.Class:
			_, err = Ssge.ParsePage(1)
		case Myhomege.Class:
			_, err = Myhomege.ParsePage(1)
		}

		if err != nil {
			log.Println("Error parsing first pages: ", err)
		}
		time.Sleep(10 * time.Minute)
	}
}

func reparseAllPages(class string) {
	defer func() {
		log.Println("reparseAllPages ended " + class)
	}()

	redisClient := NewRedisClient()
	defer redisClient.Close()

	var page uint16
	storedPage, err := redisClient.ReadKey(class + "_last_page")
	log.Println(class+"_last_page: ", storedPage)
	if err == nil {
		pageInt, err := strconv.Atoi(storedPage)
		if err == nil {
			page = uint16(pageInt)
		} else {
			page = 1
		}
	} else {
		page = 1
	}

	var delay time.Duration
	switch class {
	case "MyAutoGe":
		delay = 5 * time.Second
	case "MyAutoGeRent":
		delay = 120 * time.Second
	case Ssge.Class:
		delay = 5 * time.Second
	case Myhomege.Class:
		delay = 5 * time.Second
	}
	for {
		switch class {
		case "MyAutoGe", "MyAutoGeRent":
			page, err = Myautoge.ParsePage(page, class)
		case Ssge.Class:
			page, err = Ssge.ParsePage(page)
		case Myhomege.Class:
			page, err = Myhomege.ParsePage(page)
		}

		if err != nil {
			log.Println("Error parsing " + class + ", p " + strconv.Itoa(int(page)))
			log.Println(err)
			time.Sleep(15 * time.Second)
		}

		if page == 0 {
			page = 1

			err = redisClient.DeleteKey(class + "_last_page")
			if err != nil {
				log.Println("Error deleting last page from redis: ", err)
			}

			reparseStart, err := redisClient.ReadKey("reparse_start_" + class)
			Dbmethods.MarkAddsTrashed(class, reparseStart)
			if err != nil {
				log.Println("Error retrieve reparse_start: ", err)
			}

			err = redisClient.WriteTime("reparse_start_"+class, time.Now())
			if err != nil {
				log.Println("Error reparse_start last page to redis: ", err)
			}
		} else {
			maxPage, err := redisClient.ReadKey("max_page_" + class)
			if err != nil {
				maxPage = "0"
			}
			maxPageInteger, err := strconv.Atoi(maxPage)

			err = redisClient.WriteKey("max_page_"+class, strconv.Itoa(max(int(page+1), maxPageInteger)))
			if err != nil {
				log.Println("Error writing max_page page to redis: ", err)
			}

			err = redisClient.WriteKey(class+"_last_page", strconv.Itoa(int(page)))
			if err != nil {
				log.Println("Error writing last page to redis: ", err)
			}

			err = redisClient.WriteTime("resent_check_"+class, time.Now())
			if err != nil {
				log.Println("Error writing resent check to redis: ", err)
			}
		}

		time.Sleep(delay)
	}
}
