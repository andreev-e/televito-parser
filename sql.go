package main

import (
	"database/sql"
	"errors"
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

type User struct {
	id      int
	name    string
	contact uint32
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

func getLocationByAddress(address string, lat float32, lng float32) uint16 {
	location := QueryLocation(address)
	if location != nil {
		return 777
	}

	return CreateLocation(address).ID
}

func CreateLocation(address string) Location {
	return Location{}
}

func QueryLocation(address string) *Location {
	// Query location from database
	return nil
}

func UpdateAdd(add Add) {
	var query = "UPDATE adds SET user_id = ?, name = ?, description = ?, price = ?, price_usd = ?, currency = ?, category_id = ?, location_id = ?,  WHERE id = (?)"
	_, _ = RunQuery(query, add.user_id, add.name, add.description, add.price, add.price_usd, add.currency, add.categoryId, add.location_id, add.id)
}

func InsertAdd(add Add) {
	var query = "INSERT INTO adds (user_id, status, location_id, name, description, price, price_usd, source_class, source_id, category_id, approved, images, currency, updated_at) " +
		"VALUES (?, 2, ?, ?, ?, ?, ?, ?, ?, ?, 1, '[]', ?, NOW());"
	_, err := RunQuery(query, add.user_id, add.location_id, add.name, add.description, add.price, add.price_usd, sourceClass, add.source_id, add.categoryId, add.currency)
	panic(err)

}

func findUserByPhone(phone uint64) (User, error) {
	var user User
	var query = "SELECT * FROM users WHERE contact = ?"
	rows, err := RunQuery(query, phone)
	if err != nil {
		return user, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&user.id, &user.name, &user.contact)
		if err != nil {
			return user, err
		}

		return user, nil
	}

	return user, errors.New("user not found")
}
