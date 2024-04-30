package Dbmethods

import (
	"database/sql"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
	Lrucache "televito-parser/lrucache"
	Models "televito-parser/models"
	"time"
)

type Location struct {
	gorm.Model
	id         uint16
	lat        float32
	lng        float32
	address    string
	created_at time.Time
	updated_at time.Time
}

var db *sql.DB
var gormDb *gorm.DB

func InitDB() {
	var err error
	db, err = sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		log.Println("Error initializing database connection:")
		log.Println(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	gormDb, err = gorm.Open(mysql.Open(os.Getenv("MYSQL_CONNECTION_STRING")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func GetDbStats() sql.DBStats {
	return db.Stats()
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
	placeholders := make([]string, len(sourceIds))
	args := make([]interface{}, len(sourceIds)+1)
	for i, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
		placeholders[i] = "?"
		args[i] = sourceId
	}
	sourceIdsString = sourceIdsString[:len(sourceIdsString)-1]
	placeholdersString := strings.Join(placeholders, ",")
	args[len(args)-1] = sourceClass

	var query = "SELECT * FROM adds WHERE source_id IN (" + placeholdersString + ") AND source_class = ?"
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = make(map[uint32]Models.Add)
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

		if err == nil {
			result[add.Source_id] = add
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func RestoreTrashedAdds(sourceIds []uint32, sourceClass string) {
	var sourceIdsString string
	for _, sourceId := range sourceIds {
		sourceIdsString = sourceIdsString + strconv.Itoa(int(sourceId)) + ","
	}
	sourceIdsString = sourceIdsString[:len(sourceIdsString)-1]

	rows, err := db.Query("UPDATE adds SET deleted_at = null, updated_at = NOW() WHERE deleted_at IS NOT NULL AND source_id IN (?) AND source_class = ?", sourceIdsString, sourceClass)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
}

func GetLocationIdByAddress(address string, lat float32, lng float32) uint16 {

	locationId, err := Lrucache.CachedLocations.Get(address)
	if err == nil {
		id, err := strconv.Atoi(locationId)
		if err == nil {
			return uint16(id)
		}
	}

	var location Location

	gormDb.Unscoped().First(&location, "address = ?", address)
	if location.id != 0 {
		Lrucache.CachedLocations.Put(address, strconv.Itoa(int(location.id)))
		return location.id
	}

	gormDb.Create(&Location{address: address, lat: lat, lng: lng, created_at: time.Now(), updated_at: time.Now()})
	if location.id != 0 {
		Lrucache.CachedLocations.Put(address, strconv.Itoa(int(location.id)))
		return location.id
	}

	//location, err := queryLocation(address)
	//if err == nil {
	//	Lrucache.CachedLocations.Put(address, strconv.Itoa(int(location.id)))
	//	return location.id
	//}

	//location, err = storeLocation(address, lat, lng)
	//if err == nil {
	//	Lrucache.CachedLocations.Put(address, strconv.Itoa(int(location.id)))
	//	return location.id
	//}
	panic("location getting failed")
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
	rows, err := db.Query(query, address)
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

func UpdateAdd(add Models.Add) {
	var query = "UPDATE adds SET user_id = ?, name = ?, description = ?, price = ?, price_usd = ?, currency = ?, category_id = ?, location_id = ?, images = ? WHERE id = ?"
	rows, err := db.Query(query, add.User_id, add.Name, add.Description, add.Price, add.Price_usd, add.Currency, add.CategoryId, add.Location_id, add.Images, add.Id)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
}

func UpdateAddsBulk(adds []Models.Add) {
	if len(adds) == 0 {
		return
	}

	for _, add := range adds {
		UpdateAdd(add)
	}
}

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

	query := "INSERT INTO adds (user_id, status, location_id, name, description, price, price_usd, source_class, source_id, category_id, approved, images, currency, updated_at, created_at) VALUES " + strings.Join(valueStrings, ", ")

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(valueArgs...)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
}

func FindUserByPhone(phone string) (Models.User, error) {
	var user Models.User
	var query = "SELECT id, contact FROM users WHERE contact = ? LIMIT 1"
	rows, err := db.Query(query, phone)
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

func FindUserBySourceId(sourceId string) (Models.User, error) {
	var user Models.User
	var query = "SELECT id, contact FROM users WHERE source_id = ? LIMIT 1"
	rows, err := db.Query(query, sourceId)
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

func CreateUser(contact string, lang string, currency string, locationId uint16, sourceId interface{}) (Models.User, error) {
	var user Models.User

	if db == nil {
		return user, errors.New("database connection not initialized")
	}

	stmt, err := db.Prepare("INSERT INTO users (contact, lang, currency, location_id, created_at, updated_at, timezone, source_id) " +
		"VALUES (?,?,?,?, NOW(), NOW(), 'Asia/Tbilisi', ?);")
	if err != nil {
		return user, err
	}

	res, err := stmt.Exec(contact, lang, currency, locationId, sourceId)
	if err != nil {
		return user, err
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return user, err
	}

	user.Id = int(userId)
	user.Contact = contact
	user.Lang = lang
	user.Currency = currency
	user.Location_id = locationId

	return user, nil
}

func FindCategoryByNameAndParent(name string, parentId uint16) (Models.Category, error) {
	var category Models.Category
	var query = "SELECT id FROM categories WHERE name = ? AND parent_id = ? AND deleted_at IS NULL"
	rows, err := db.Query(query, name, parentId)
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

	categoryId, err := res.LastInsertId()
	if err != nil {
		return category, err
	}

	category.Id = uint16(int(categoryId))
	category.Name = name
	category.ParentId = parentId

	return category, nil
}

func MarkAddsTrashed(sourceClass string, olderThan string) {
	rows, err := db.Query("UPDATE adds SET deleted_at = NOW() WHERE deleted_at IS NULL AND source_class = ? AND updated_at < ?", sourceClass, olderThan)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
}
