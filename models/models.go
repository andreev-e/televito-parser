package Models

import "gorm.io/gorm"

type Add struct {
	gorm.Model
	Id           uint64
	User_id      int
	Status       int
	Location_id  uint64
	Name         string
	Description  string
	Price        int
	Price_usd    float32
	Source_class string
	Source_id    uint64
	CategoryId   uint64
	Approved     int
	Images       string
	Currency     string
	Updated_at   string
	Created_at   string
	Deleted_at   *string
}

type Location struct {
	gorm.Model
	ID      uint64 `gorm:"primaryKey"`
	Lat     float32
	Lng     float32
	Address string
}

type Category struct {
	gorm.Model
	Id         uint64 `gorm:"primaryKey"`
	Name       string
	ParentId   uint64
	Created_at string
	Updated_at string
	Deleted_at string
	Adds_count uint32
}

type User struct {
	Id          int
	Contact     interface{}
	Lang        string
	Currency    string
	Location_id uint64
	Timezone    string
	Created_at  string
	Updated_at  string
}
