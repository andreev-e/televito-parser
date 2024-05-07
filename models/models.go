package Models

import (
	"gorm.io/gorm"
)

type Add struct {
	gorm.Model
	UserId          uint
	Status          int
	Location_id     uint64
	Name            string
	Description     string
	Price           float32
	Price_usd       float32
	Source_class    string
	Source_id       uint64
	CategoryId      uint
	Approved        int
	Images          string
	Currency        string
	Characteristics []Characteristic
}

type Location struct {
	gorm.Model
	Lat     float32
	Lng     float32
	Address string
}

type Category struct {
	gorm.Model
	Name       string
	ParentId   uint
	Adds_count uint32
}

type User struct {
	gorm.Model
	Contact    interface{}
	Lang       string
	Currency   string
	LocationId uint64
	Location   Location
	Timezone   string
}

type Characteristic struct {
	gorm.Model
	AddId uint
	Add   Add
	Class string
	Value string
}
