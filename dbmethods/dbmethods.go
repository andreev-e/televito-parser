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

func FirstOrCreate(add Models.Add) {
	var existingAdd Models.Add
	result := gormDb.Where(Models.Add{Source_id: add.Source_id, Source_class: add.Source_class, Deleted_at: nil}).First(&existingAdd)

	if result.Error == nil {
		gormDb.Model(&existingAdd).Updates(add)
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		gormDb.Create(&add)
	} else {
		// Handle other errors if necessary
	}
}
