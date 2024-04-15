package main

import "time"

func reparseFirstPages() {
	for {
		MyAutoGeParsePage(1)
		time.Sleep(10 * time.Minute)
	}
}

func reparseAllPages() {
	var page uint16
	page = 1
	for {
		page = MyAutoGeParsePage(page)
		time.Sleep(10 * time.Second)
	}
}
