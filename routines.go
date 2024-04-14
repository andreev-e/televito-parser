package main

import "time"

func reparseFirstPages() {
	for {
		MyAutoGeParsePage(1)
		time.Sleep(10 * time.Minute)
	}
}
