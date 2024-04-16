package main

import (
	"strconv"
	"time"
)

func reparseFirstPages() {
	for {
		MyAutoGeParsePage(1)
		time.Sleep(10 * time.Minute)
	}
}

func reparseAllPages() {
	var page uint16
	page = 1
	storedPage, err := readRedisKey("tvito_database_tvito_cache_:MyAutoGe_last_page")
	if err != nil {
		pageInt, err := strconv.Atoi(storedPage)
		if err != nil {
			page = uint16(pageInt)
		} else {
			page = 777
		}
	}

	for {
		page = MyAutoGeParsePage(page)
		time.Sleep(1 * time.Second)
	}
}
