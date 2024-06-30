package Myautoge

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	Consts "televito-parser/consts"
	"televito-parser/dbmethods"
	Main "televito-parser/models"
	"time"
)

var (
	once sync.Once
)

type AddSource struct {
	CarID         uint64  `json:"car_id"`
	Price         float32 `json:"price"`
	PriceUSD      float32 `json:"price_usd"`
	Currency      uint8   `json:"currency_id"`
	ManID         uint16  `json:"man_id"`
	ModelID       uint16  `json:"model_id"`
	CarModel      string  `json:"car_model"`
	CarDesc       string  `json:"car_desc"`
	ProdYear      uint16  `json:"prod_year"`
	EngineVolume  uint16  `json:"engine_volume"`
	GearTypeID    uint16  `json:"gear_type_id"`
	VehicleType   uint16  `json:"vehicle_type"`
	CustomsPassed bool    `json:"customs_passed"`
	LocationId    uint16  `json:"location_id"`
	ClientPhone   string  `json:"client_phone"`
	PhotosCount   uint    `json:"pic_number"`
	Photo         string  `json:"photo"`
	CategoryID    uint16  `json:"category_id"`
	CarRunKm      uint64  `json:"car_run_km"`
	FuelTypeId    uint16  `json:"fuel_type_id"`
}

type Response struct {
	Data struct {
		AddSources []AddSource `json:"items"`
	} `json:"data"`
}

type LoadedAppData struct {
	Categories []Category     `json:"Categories"`
	Currencies []Currency     `json:"Currencies"`
	Models     []Model        `json:"Models"`
	Mans       []Manufacturer `json:"Mans"`
	GearTypes  struct {
		Items []GearType `json:"items"`
	} `json:"GearTypes"`
	Locations struct {
		Items []Location `json:"items"`
	} `json:"Locations"`
	FuelTypes struct {
		Items []FuelType `json:"items"`
	} `json:"FuelTypes"`
}

type AppData struct {
	Categories map[uint16]Category
	Currencies map[uint8]Currency
	Models     map[uint16]Model
	Mans       map[uint16]Manufacturer
	Locations  map[uint16]Location
	GearTypes  map[uint16]GearType
	FuelTypes  map[uint16]FuelType
}

type Category struct {
	CategoryId uint16 `json:"category_id"`
	Name       string `json:"title"`
}

type Currency struct {
	ID     uint8  `json:"currencyID"`
	Name   string `json:"title"`
	Symbol string `json:"currencySymbol"`
}

type Model struct {
	ID   uint16 `json:"model_id"`
	Name string `json:"model_name"`
}

type Manufacturer struct {
	ID   uint16 `json:"man_id"`
	Name string `json:"man_name"`
}

type Location struct {
	ID       uint16 `json:"location_id"`
	Name     string `json:"title"`
	ParentId uint16 `json:"parent_id"`
}

type GearType struct {
	ID   uint16 `json:"gear_type_id"`
	Name string `json:"title"`
}

type FuelType struct {
	ID   uint16 `json:"fuel_type_id"`
	Name string `json:"title"`
}

const url = "https://api2.myauto.ge"
const NumberOfPhotos uint = 5
const mainCategory = 12

var Characteristics = []string{
	Consts.Mileage,
	Consts.ProductionYear,
	Consts.VehicleBodyType,
	Consts.FuelType,
	Consts.TransmissionType,
}

var autoAppData AppData

func LoadPage(page uint16, class string) ([]Main.Add, error) {
	loadData()
	var forRent = "0"
	if class == "MyAutoGeRent" {
		forRent = "1"
	}

	params := map[string]string{
		"ForRent":       forRent,
		"CurrencyID":    "1",
		"MileageType":   "1",
		"SortOrder":     "1",
		"Page":          strconv.Itoa(int(page)),
		"hideDealPrice": "1",
		"Locs":          "2.3.4.7.15.30.113.52.37.36.38.39.40.31.5.41.44.47.48.53.54.8.16.6.14.13.12.11.10.9.55.56.57.59.58.61.62.63.64.66.71.72.74.75.76.77.78.80.81.82.83.84.85.86.87.88.91.96.97.101.109",
	}

	fullUrl := url + "/ka/products/?"
	for key, value := range params {
		fullUrl += key + "=" + value + "&"
	}

	response, err := http.Get(fullUrl)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var responseObject Response
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}

	result := make([]Main.Add, 0)

	for _, addSource := range responseObject.Data.AddSources {
		address := getAddress(addSource.LocationId, "")
		locationId := Dbmethods.GetLocationIdByAddress(address, 0, 0)
		user, userError := getUser(addSource, locationId)
		category, categoryError := getCategory(addSource)

		if userError == nil && categoryError == nil {
			var add Main.Add
			add.UserId = user.ID
			add.Status = 2
			add.Approved = 1
			add.Location_id = locationId
			add.Name = getName(addSource)
			add.Description = getDescription(addSource)
			add.Price = addSource.Price
			add.Price_usd = addSource.PriceUSD
			add.Source_class = class
			add.Source_id = addSource.CarID
			add.CategoryId = category.ID
			add.Images = getImagesUrlList(addSource, addSource.CarID)
			add.Currency = getCurrency(addSource)
			add.UpdatedAt = time.Now()

			add.Characteristics = getCharacteristics(addSource)

			result = append(result, add)
		}

	}
	return result, nil
}

func getCharacteristics(source AddSource) []Main.Characteristic {
	result := make([]Main.Characteristic, 0)
	for _, characteristic := range Characteristics {
		value := ""

		switch characteristic {
		case Consts.Mileage:
			value = strconv.FormatUint(source.CarRunKm, 10)
		case Consts.ProductionYear:
			value = strconv.Itoa(int(source.ProdYear))
		case Consts.VehicleBodyType:
			category, ok := autoAppData.Categories[source.CategoryID]
			if ok {
				value = category.Name
			}
		case Consts.FuelType:
			fuel, ok := autoAppData.FuelTypes[source.FuelTypeId]
			if ok {
				log.Println(fuel.Name)
				value = fuel.Name
			}
		}

		if value != "" {
			char := Main.Characteristic{
				Class: characteristic,
				Value: value,
			}

			result = append(result, char)
		}
	}

	return result
}

func getImagesUrlList(addSource AddSource, id uint64) string {
	images := make([]string, 0)
	for i := uint(1); i <= min(addSource.PhotosCount, NumberOfPhotos); i++ {
		url := "https://static.my.ge/myauto/photos/" + addSource.Photo + "/large/" + strconv.Itoa(int(id)) + "_" + strconv.Itoa(int(i)) + ".jpg"
		images = append(images, strings.ReplaceAll(url, "/", "\\/"))
	}

	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

func getUser(addSource AddSource, locationId uint64) (Main.User, error) {
	var phone = addSource.ClientPhone
	var user, err = Dbmethods.FindUserByPhone(phone)
	if err != nil {
		user, err = Dbmethods.CreateUser(phone, "ge", getCurrency(addSource), locationId, nil)
	}
	return user, err
}

func getAddress(locationId uint16, address string) string {
	location, ok := autoAppData.Locations[locationId]
	if ok {
		address = address + location.Name + ", "
		if location.ParentId != 0 {
			return getAddress(location.ParentId, address)
		}
	} else {
		log.Println(autoAppData.Locations)
		panic("No location in dictionary " + strconv.Itoa(int(locationId)))
	}

	return address[:len(address)-2]
}

func getCategory(addSource AddSource) (Main.Category, error) {
	manufacturer, manufacturerOk := autoAppData.Mans[addSource.ManID]
	if !manufacturerOk {
		return Main.Category{}, fmt.Errorf("Manufacturer not found")
	}

	var category Main.Category
	var subCategory Main.Category
	var err = fmt.Errorf("Get category error")

	createdCategory, err := Dbmethods.RetrieveCategory(manufacturer.Name, mainCategory)
	if err != nil {
		return Main.Category{}, err
	}
	category = createdCategory

	subCategoryAuto, subCatOk := autoAppData.Categories[addSource.CategoryID]
	if !subCatOk {
		return category, nil
	}

	subCategory, err = Dbmethods.RetrieveCategory(subCategoryAuto.Name, category.ID)
	if err != nil {
		return category, err
	}

	return subCategory, nil
}

func getCurrency(addSource AddSource) string {
	currency, ok := autoAppData.Currencies[addSource.Currency]
	if ok {
		if currency.Name == "EURO" {
			return "EUR"
		}

		return currency.Name
	}
	return ""
}

func getDescription(addSource AddSource) string {
	var description []string

	if addSource.CarDesc != "" {
		description = append(description, addSource.CarDesc)
	}

	if !addSource.CustomsPassed {
		description = append(description, "🚫 Customs not passed")
	}

	return strings.Join(description, "\n\r")
}

func getName(addSource AddSource) string {
	var name []string

	manufacturer, ok := autoAppData.Mans[addSource.ManID]
	if ok {
		name = append(name, manufacturer.Name)
	}

	model, ok := autoAppData.Models[addSource.ModelID]
	if ok {
		name = append(name, model.Name)
	}

	if addSource.CarModel != "" {
		name = append(name, addSource.CarModel)
	}

	if addSource.ProdYear != 0 {
		name = append(name, strconv.Itoa(int(addSource.ProdYear)))
	}

	gearType, transmissionOk := autoAppData.GearTypes[addSource.GearTypeID]

	if ok && addSource.EngineVolume != 0 {
		var transmission = ""

		if transmissionOk {
			switch gearType.Name {
			case "Automatic":
				transmission = "AT"
				break
			case "Tiptronic":
				transmission = "AT"
				break
			case "Manual":
				transmission = "MT"
				break
			case "Variator":
				transmission = "CVT"
				break
			}
		}

		if addSource.VehicleType == 2 {
			name = append(name, strconv.Itoa(int(addSource.EngineVolume)))
		} else {
			name = append(name, strconv.FormatFloat(float64(addSource.EngineVolume)/1000, 'f', 1, 64)+transmission)
		}

	}

	return strings.Join(name, " ")
}

func loadData() {
	once.Do(func() {
		response, err := http.Get(url + "/appdata/other_en.json")

		log.Println(url + "/appdata/other_en.json")
		if err != nil {
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(response.Body)

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return
		}

		var loadedAppData LoadedAppData

		err = json.Unmarshal(body, &loadedAppData)
		if err != nil {
			log.Printf("Error parsing JSON: %v\n", err)
			panic("Can't load myAutoGe appdata")
		}

		categories := make(map[uint16]Category)
		for _, category := range loadedAppData.Categories {
			categories[category.CategoryId] = category
		}

		mans := make(map[uint16]Manufacturer)
		for _, manufacturer := range loadedAppData.Mans {
			mans[manufacturer.ID] = manufacturer
		}

		locations := make(map[uint16]Location)
		for _, location := range loadedAppData.Locations.Items {
			locations[location.ID] = location
		}

		models := make(map[uint16]Model)
		for _, model := range loadedAppData.Models {
			models[model.ID] = model
		}

		gearTypes := make(map[uint16]GearType)
		for _, gearType := range loadedAppData.GearTypes.Items {
			gearTypes[gearType.ID] = gearType
		}

		fuelTypes := make(map[uint16]FuelType)
		for _, fuelType := range loadedAppData.FuelTypes.Items {
			fuelTypes[fuelType.ID] = fuelType
		}

		currencies := make(map[uint8]Currency)
		for _, currency := range loadedAppData.Currencies {
			currencies[currency.ID] = currency
		}

		autoAppData = AppData{
			Categories: categories,
			Mans:       mans,
			Locations:  locations,
			Models:     models,
			GearTypes:  gearTypes,
			Currencies: currencies,
			FuelTypes:  fuelTypes,
		}

		log.Println("Appdata loaded")
	})
}
