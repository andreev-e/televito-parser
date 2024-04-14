package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Add struct {
	id           int
	user_id      int
	status       int
	location_id  int
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
