package main

import (
	"fmt"
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
	storedPage, err := readRedisKey("tvito_database_tvito_cache_:MyAutoGe_last_page")
	fmt.Println("tvito_database_tvito_cache_:MyAutoGe_last_page: ", storedPage)
	if err == nil {
		pageInt, err := strconv.Atoi(storedPage)
		if err != nil {
			page = uint16(pageInt)
		} else {
			page = 777
		}
	} else {
		page = 1
	}

	for {
		page, err = MyAutoGeParsePage(page)
		time.Sleep(5 * time.Second)
	}
}
