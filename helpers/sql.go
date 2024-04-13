package helpers

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
)

type Add struct {
	id          int
	userId      int
	status      int
	locationId  int
	name        string
	description string
	price       int
	priceUsd    float32
	sourceClass string
	sourceId    int
	categoryId  int
	approved    int
	images      string
	currency    string
}

func WriteAdds(adds []Add) {

}

func GetExistingAdds(sourceIds []int32) []Add {

	db, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var sourceIdsString string
	for _, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
	}

	fmt.Println(sourceIdsString)

	rows, err := db.Query("SELECT id, name FROM adds WHERE source_id IN (?)", sourceIdsString)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	result := make([]Add, 0)
	for rows.Next() {
		var add Add

		//scn rows into Add struct

		err := rows.Scan(&add.id, &add.name)
		if err != nil {
			panic(err)
		}

		result = append(result, add)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return result
}
