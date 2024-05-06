package Routines

import (
	"log"
	"strconv"
	Myautoge "televito-parser/addsources/myautoge"
	Myhomege "televito-parser/addsources/myhomege"
	Ssge "televito-parser/addsources/ssge"
	Dbmethods "televito-parser/dbmethods"
	Models "televito-parser/models"
	Redis "televito-parser/redis"
	"time"
)

func ReparseFirstPages(class string) {
	defer func() {
		log.Println("reparseFirstPages ended " + class)
	}()

	for {
		err := error(nil)
		switch class {
		case "MyAutoGe", "MyAutoGeRent":
			_, err = Myautoge.LoadPage(1, class)
		case Ssge.Class:
			_, err = Ssge.LoadPage(1, class)
		case Myhomege.Class:
			_, err = Myhomege.LoadPage(1, class)
		}
		if err != nil {
			log.Println("Error parsing first pages: ", err)
		}

		time.Sleep(5 * time.Minute)
	}
}

func ReparseAllPages(class string) {
	defer func() {
		log.Println("reparseAllPages ended " + class)
	}()

	redisClient := Redis.NewRedisClient()
	defer redisClient.Close()

	var page uint16
	storedPage, err := redisClient.ReadKey(class + "_last_page")
	log.Println(class+"_last_page: ", storedPage)
	if err == nil {
		pageInt, err := strconv.Atoi(storedPage)
		if err == nil && pageInt > 0 {
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
		adds := make([]Models.Add, 0)
		switch class {
		case "MyAutoGe", "MyAutoGeRent":
			adds, err = Myautoge.LoadPage(page, class)
		case Ssge.Class:
			adds, err = Ssge.LoadPage(page, class)
		case Myhomege.Class:
			adds, err = Myhomege.LoadPage(page, class)
		}

		if err != nil {
			log.Println(class + " Error parsing, p " + strconv.Itoa(int(page)))
			log.Println(err)
			time.Sleep(2 * time.Minute)
		}

		if (len(adds)) == 0 {
			page = 1

			err = redisClient.DeleteKey(class + "_last_page")
			if err != nil {
				log.Println("Error deleting last page from redis: ", err)
			}

			//reparseStart, err := redisClient.ReadTime("reparse_start_" + class)
			//Dbmethods.MarkAddsTrashed(class, reparseStart)

			if err != nil {
				log.Println("Error retrieve reparse_start: ", err)
			}

			err = redisClient.WriteTime("reparse_start_"+class, time.Now())
			if err != nil {
				log.Println("Error reparse_start last page to redis: ", err)
			}
		} else {
			created, updated, errored := 0, 0, 0
			for _, add := range adds {
				result, err := Dbmethods.FirstOrCreate(add)
				if err != nil {
					log.Println("Error creating add: ", err)
					errored++
				}
				if result {
					created++
				} else {
					updated++
				}
			}
			log.Println(class, " created: ", created, " updated: ", updated, " errored: ", errored)

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

			page++
		}

		time.Sleep(delay)
	}
}