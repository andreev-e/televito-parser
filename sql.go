package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strconv"
)

type Add struct {
	id           int
	user_id      int
	status       int
	location_id  uint16
	name         string
	description  string
	price        int
	price_usd    float32
	source_class string
	source_id    uint32
	categoryId   int
	approved     int
	images       string
	currency     string
	updated_at   string
}

func GetExistingAdds(sourceIds []uint32, source_class string) map[uint32]Add {
	var sourceIdsString string
	for _, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
	}

	rows, _ := RunQuery("SELECT * FROM adds WHERE source_id IN (?) AND source_class = (?)", sourceIdsString, source_class)

	result := make(map[uint32]Add, 0)
	for rows.Next() {
		var add Add

		err := rows.Scan(&add.id, &add.user_id, &add.source_id)
		if err != nil {
			panic(err)
		}

		result[add.source_id] = add
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return result
}

func RestoreTrashedAdds(sourceIds []uint32, sourceClass string) {
	var sourceIdsString string
	for _, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
	}

	_, _ = RunQuery("UPDATE adds SET deleted_at = null, updated_at = NOW() WHERE deleted_at IS NOT NULL AND source_id IN (?) AND source_class = (?)", sourceIdsString, sourceClass)
	fmt.Println("trashed restored")
}

func RunQuery(query string, params ...interface{}) (*sql.Rows, error) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func getLocationByAddress(address []rune, lat float32, lng float32) uint16 {
	var transliterated = Transliterate(address[:150])
	panic(transliterated)
	// Query the location by address
	location := QueryLocation(transliterated)
	if location != nil {
		return 777
	}

	return CreateLocation(transliterated).ID
}

func CreateLocation(address string) Location {
	return Location{}
}

func QueryLocation(address string) *Location {
	// Query location from database
	return nil
}
