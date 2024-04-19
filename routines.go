package main

import (
	"log"
	"strconv"
	"televito-parser/addsources/myautoge"
	Ssge "televito-parser/addsources/ssge"
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
	storedPage, err := redisClient.ReadKey("tvito_database_tvito_cache_:" + class + "_last_page")
	log.Println("tvito_database_tvito_cache_:"+class+"_last_page: ", storedPage)
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

	err = redisClient.WriteKey("_tvito_database_tvito_cache_:"+class+"_last_page", strconv.Itoa(int(page)))
	if err != nil {
		log.Println("Error writing last page to redis: ", err)
	}

	var delay time.Duration
	switch class {
	case "MyAutoGe":
		delay = 5 * time.Second
	case "MyAutoGeRent":
		delay = 20 * time.Second
	case Ssge.Class:
		delay = 20 * time.Second
	}
	for {
		switch class {
		case "MyAutoGe", "MyAutoGeRent":
			page, err = Myautoge.ParsePage(page, class)
		case Ssge.Class:
			page, err = Ssge.ParsePage(page)
		}

		if err != nil {
			log.Println("Error parsing "+class+", p "+strconv.Itoa(int(page)), err)
			time.Sleep(60 * time.Second)
		}
		time.Sleep(delay)
	}
}
