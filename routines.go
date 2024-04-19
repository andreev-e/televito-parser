package main

import (
	"log"
	"strconv"
	Myautoge "televito-parser/myautoge"
	"time"
)

func reparseFirstPages(class string) {
	defer func() {
		log.Println("reparseFirstPages ended " + class)
	}()

	for {
		_, err := Myautoge.ParsePage(1, class)
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

	var page uint16
	storedPage, err := readRedisKey("tvito_database_tvito_cache_:" + class + "_last_page")
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

	var delay time.Duration
	switch class {
	case "MyAutoGe":
		delay = 5 * time.Second
	case "MyAutoGeRent":
		delay = 20 * time.Second
	}
	for {
		page, err = Myautoge.ParsePage(page, class)
		if err != nil {
			log.Println("Error parsing "+class+", p "+strconv.Itoa(int(page)), err)
			time.Sleep(60 * time.Second)
		}
		time.Sleep(delay)
	}
}
