package main

import (
	"fmt"
	"strconv"
	Myautoge "televito-parser/myautoge"
	"time"
)

func reparseFirstPages() {
	for {
		_, err := Myautoge.MyAutoGeParsePage(1)
		if err != nil {
			fmt.Println("Error parsing first pages: ", err)
		}
		time.Sleep(10 * time.Minute)
	}
}

func reparseAllPages(class string) {
	var page uint16
	storedPage, err := readRedisKey("tvito_database_tvito_cache_:" + class + "_last_page")
	fmt.Println("tvito_database_tvito_cache_:"+class+"_last_page: ", storedPage)
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

	for {
		page, err = Myautoge.MyAutoGeParsePage(page)
		time.Sleep(10 * time.Second)
	}
}
