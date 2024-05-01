package Dbmethods

import (
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
	Lrucache "televito-parser/lrucache"
	Models "televito-parser/models"
)

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

	gormDb, err = gorm.Open(mysql.Open(os.Getenv("MYSQL_CONNECTION_STRING")+"?parseTime=true"), &gorm.Config{})
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

func GetLocationIdByAddress(address string, lat float32, lng float32) uint64 {
	locationId, err := Lrucache.CachedLocations.Get(address)
	if err == nil {
		id, err := strconv.ParseUint(locationId, 10, 64)
		if err == nil {
			return id
		}
	}

	var location Models.Location

	if err := gormDb.First(&location, "address = ?", address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := gormDb.Create(&Models.Location{Address: address, Lat: lat, Lng: lng}).Error; err != nil {
				return 0
			}

			if err := gormDb.First(&location, "address = ?", address).Error; err != nil {
				return 0
			}
			Lrucache.CachedLocations.Put(address, strconv.FormatUint(location.ID, 10))
			return location.ID
		}
		return 0
	}

	Lrucache.CachedLocations.Put(address, strconv.FormatUint(location.ID, 10))
	return location.ID
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

func CreateUser(contact string, lang string, currency string, locationId uint64, sourceId interface{}) (Models.User, error) {
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

func RetrieveCategory(name string, parentId uint64) (Models.Category, error) {
	var category Models.Category

	if err := gormDb.First(&category, "name = ? AND parent_id = ?", name, parentId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := gormDb.Create(&Models.Category{Name: name, ParentId: parentId}).Error; err != nil {
				return category, fmt.Errorf("failed to create category: %w", err)
			}
			if err := gormDb.First(&category, "name = ? AND parent_id = ?", name, parentId).Error; err != nil {
				return category, fmt.Errorf("failed to retrieve created category: %w", err)
			}
			return category, nil
		}
		return category, fmt.Errorf("failed to retrieve category: %w", err)
	}

	return category, nil
}

func MarkAddsTrashed(sourceClass string, olderThan string) {
	gormDb.Model(&Models.Add{}).
		Where("deleted_at IS NULL AND source_class = ? AND updated_at < ?", sourceClass, olderThan).
		Updates(map[string]interface{}{"deleted_at": olderThan})
}
