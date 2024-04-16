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
	categoryId   uint16
	approved     int
	images       string
	currency     string
	updated_at   string
	created_at   string
	deleted_at   *string
}

type User struct {
	id          int
	contact     uint64
	lang        string
	currency    string
	location_id uint16
	timezone    string
	created_at  string
	updated_at  string
}

func GetExistingAdds(sourceIds []uint32, source_class string) (map[uint32]Add, error) {
	var sourceIdsString string
	for _, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
	}

	rows, err := RunQuery("SELECT * FROM adds WHERE source_id IN (?) AND source_class = (?)", sourceIdsString, source_class)
	if err != nil {
		return nil, err
	}

	result := make(map[uint32]Add, 0)
	for rows.Next() {
		var add Add

		err := rows.Scan(
			&add.id,
			&add.user_id,
			&add.name,
			&add.description,
			&add.price,
			&add.price_usd,
			&add.currency,
			&add.images,
			&add.categoryId,
			&add.location_id,
			&add.status,
			&add.approved,
			&add.created_at,
			&add.updated_at,
			&add.deleted_at,
			&add.source_class,
			&add.source_id,
		)

		if err != nil {
			return result, err
		}

		result[add.source_id] = add
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
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
	location, err := queryLocation(address)
	if err == nil {
		return location.id
	}

	location, err = storeLocation(address, lat, lng)
	if err == nil {
		return location.id
	}
	panic("location store failed")
}

type Location struct {
	id         uint16
	address    string
	created_at string
	updated_at string
}

func storeLocation(address string, lat float32, lng float32) (Location, error) {
	var location Location

	db, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		return location, err
	}
	defer db.Close()

	stmt, _ := db.Prepare("INSERT INTO locations (address, lat, lng, created_at, updated_at) " +
		"VALUES (?,?,?, NOW(), NOW());")
	defer stmt.Close()

	res, err := stmt.Exec(address, lat, lng)
	if err != nil {
		return location, err
	}

	locationId, _ := res.LastInsertId()

	location.id = uint16(int(locationId))
	location.address = address

	return location, nil
}

func queryLocation(address string) (Location, error) {
	var query = "SELECT id, address FROM locations WHERE address = ?"
	rows, err := RunQuery(query, address)
	if err != nil {
		return Location{}, err
	}
	defer rows.Close()

	if err == nil {
		for rows.Next() {
			var location Location
			_ = rows.Scan(
				&location.id,
				&location.address,
			)

			return location, nil
		}
	}

	return storeLocation(address, 0, 0)
}

func UpdateAdd(add Add) {
	var query = "UPDATE adds SET user_id = ?, name = ?, description = ?, price = ?, price_usd = ?, currency = ?, category_id = ?, location_id = ?, images = ? WHERE id = ?"
	_, err := RunQuery(query, add.user_id, add.name, add.description, add.price, add.price_usd, add.currency, add.categoryId, add.location_id, add.images, add.id)
	if err != nil {
		fmt.Println(add.id)
		panic(err)
	}
}

func InsertAdd(add Add) {
	var query = "INSERT INTO adds (user_id, status, location_id, name, description, price, price_usd, source_class, source_id, category_id, approved, images, currency, updated_at, created_at) " +
		"VALUES (?, 2, ?, ?, ?, ?, ?, ?, ?, ?, 1, ? , ?, NOW(), NOW());"
	_, _ = RunQuery(query, add.user_id, add.location_id, add.name, add.description, add.price, add.price_usd, sourceClass, add.source_id, add.categoryId, add.images, add.currency)
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
		err := rows.Scan(&user.id, &user.contact)
		if err != nil {
			return user, err
		}

		return user, nil
	}

	return user, errors.New("user not found")
}

func createUser(contact uint64, lang string, currency string, locationId uint16) (User, error) {
	var user User

	db, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		return user, err
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO users (contact, lang, currency, location_id, created_at, updated_at, timezone) " +
		"VALUES (?,?,?,?, NOW(), NOW(), 'Asia/Tbilisi');")
	if err != nil {
		return user, err
	}

	res, err := stmt.Exec(contact, lang, currency, locationId)
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	userId, _ := res.LastInsertId()

	user.id = int(userId)
	user.contact = contact
	user.lang = lang
	user.currency = currency
	user.location_id = locationId

	return user, nil
}

func findCategoryByNameAndParent(name string, parentId uint16) (Category, error) {
	var category Category
	var query = "SELECT * FROM categories WHERE contact = ? AND parent_id = ? AND deleted_at IS NULL"
	rows, err := RunQuery(query, name, parentId)
	if err != nil {
		return category, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&category.id)
		if err != nil {
			return category, err
		}

		return category, nil
	}

	return category, errors.New("category not found")
}

func createCategory(name string, parentId uint16) (Category, error) {
	var category Category

	db, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		return category, err
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO categories (name, parent_id, created_at, updated_at) " +
		"VALUES (?,?, NOW(), NOW());")

	if err != nil {
		return category, err
	}

	res, err := stmt.Exec(name, parentId)
	if err != nil {
		return category, err
	}
	defer stmt.Close()

	categoryId, err := res.LastInsertId()
	if err != nil {
		return category, err
	}

	category.id = uint16(int(categoryId))
	category.name = name
	category.parentId = parentId

	return category, nil
}