package Dbmethods

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
	"strings"
	Models "televito-parser/models"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		log.Println("Error initializing database connection:")
		log.Println(err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
}

func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Println("Error closing database connection:")
			log.Println(err)
			return
		}
	}
}

func GetExistingAdds(sourceIds []uint32, sourceClass string) (map[uint32]Models.Add, error) {
	var sourceIdsString string
	for _, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
	}

	rows, err := RunQuery("SELECT * FROM adds WHERE source_id IN (?) AND source_class = (?)", sourceIdsString, sourceClass)
	if err != nil {
		return nil, err
	}

	result := make(map[uint32]Models.Add, 0)
	for rows.Next() {
		var add Models.Add

		err := rows.Scan(
			&add.Id,
			&add.User_id,
			&add.Name,
			&add.Description,
			&add.Price,
			&add.Price_usd,
			&add.Currency,
			&add.Images,
			&add.CategoryId,
			&add.Location_id,
			&add.Status,
			&add.Approved,
			&add.Created_at,
			&add.Updated_at,
			&add.Deleted_at,
			&add.Source_class,
			&add.Source_id,
		)

		if err != nil {
			return result, err
		}

		result[add.Source_id] = add
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
}

func RunQuery(query string, params ...interface{}) (*sql.Rows, error) {
	if db == nil {
		return nil, errors.New("database connection not initialized")
	}

	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func GetLocationByAddress(address string, lat float32, lng float32) uint16 {
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

	if db == nil {
		return location, errors.New("database connection not initialized")
	}

	stmt, err := db.Prepare("INSERT INTO locations (address, lat, lng, created_at, updated_at) " +
		"VALUES (?,?,?, NOW(), NOW());")
	if err != nil {
		return location, err
	}
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

	for rows.Next() {
		var location Location
		_ = rows.Scan(
			&location.id,
			&location.address,
		)

		return location, nil
	}

	return storeLocation(address, 0, 0)
}

//func UpdateAdd(add Add) {
//	var query = "UPDATE adds SET user_id = ?, name = ?, description = ?, price = ?, price_usd = ?, currency = ?, category_id = ?, location_id = ?, images = ? WHERE id = ?"
//	_, err := RunQuery(query, add.user_id, add.name, add.description, add.price, add.price_usd, add.currency, add.categoryId, add.location_id, add.images, add.id)
//	if err != nil {
//		log.Println(add.id)
//	}
//}

func UpdateAddsBulk(adds []Models.Add) {
	log.Println("Bulk Items adding" + strconv.Itoa(len(adds)))

	if len(adds) == 0 {
		return
	}

	var valueStrings []string
	var valueArgs []interface{}

	for _, add := range adds {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, add.User_id, add.Name, add.Description, add.Price, add.Price_usd, add.Currency, add.CategoryId, add.Location_id, add.Images, add.Id)
	}

	// Construct the query with multiple value strings
	query := "UPDATE adds SET user_id = ?, name = ?, description = ?, price = ?, price_usd = ?, currency = ?, category_id = ?, location_id = ?, images = ? WHERE id = ?"

	// Execute the batch insert query
	_, err := RunQuery(query, valueArgs...)
	if err != nil {
		log.Println(err)
	}
}

//func InsertAdd(add Add) {
//	var query = "INSERT INTO adds (user_id, status, location_id, name, description, price, price_usd, source_class, source_id, category_id, approved, images, currency, updated_at, created_at) " +
//		"VALUES (?, 2, ?, ?, ?, ?, ?, ?, ?, ?, 1, ? , ?, NOW(), NOW());"
//	_, _ = RunQuery(query, add.user_id, add.location_id, add.name, add.description, add.price, add.price_usd, sourceClass, add.source_id, add.categoryId, add.images, add.currency)
//}

func InsertAddsBulk(adds []Models.Add) {
	if len(adds) == 0 {
		return
	}

	var valueStrings []string
	var valueArgs []interface{}

	for _, add := range adds {
		valueStrings = append(valueStrings, "(?, 2, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, NOW(), NOW())")

		valueArgs = append(valueArgs, add.User_id, add.Location_id, add.Name, add.Description, add.Price, add.Price_usd, add.Source_class, add.Source_id, add.CategoryId, add.Images, add.Currency)
	}

	// Construct the query with multiple value strings
	query := "INSERT INTO adds (user_id, status, location_id, name, description, price, price_usd, source_class, source_id, category_id, approved, images, currency, updated_at, created_at) VALUES " + strings.Join(valueStrings, ", ")

	// Execute the batch insert query
	_, err := RunQuery(query, valueArgs...)
	if err != nil {
		log.Println(err)
	}
}

func FindUserByPhone(phone uint64) (Models.User, error) {
	var user Models.User
	var query = "SELECT * FROM users WHERE contact = ?"
	rows, err := RunQuery(query, phone)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Contact)
		if err != nil {
			return user, err
		}

		return user, nil
	}

	return user, errors.New("user not found")
}

func CreateUser(contact uint64, lang string, currency string, locationId uint16) (Models.User, error) {
	var user Models.User

	if db == nil {
		return user, errors.New("database connection not initialized")
	}

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

	user.Id = int(userId)
	user.Contact = contact
	user.Lang = lang
	user.Currency = currency
	user.Location_id = locationId

	return user, nil
}

func FindCategoryByNameAndParent(name string, parentId uint16) (Models.Category, error) {
	var category Models.Category
	var query = "SELECT * FROM categories WHERE contact = ? AND parent_id = ? AND deleted_at IS NULL"
	rows, err := RunQuery(query, name, parentId)
	if err != nil {
		return category, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&category.Id)
		if err != nil {
			return category, err
		}

		return category, nil
	}

	return category, errors.New("category not found")
}

func CreateCategory(name string, parentId uint16) (Models.Category, error) {
	var category Models.Category

	if db == nil {
		return category, errors.New("database connection not initialized")
	}

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

	category.Id = uint16(int(categoryId))
	category.Name = name
	category.ParentId = parentId

	return category, nil
}
