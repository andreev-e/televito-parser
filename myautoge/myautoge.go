package Myautoge

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"televito-parser/dbmethods"
	Http "televito-parser/http"
	Main "televito-parser/models"
)

var (
	once        sync.Once
	loadedMutex sync.Mutex
)

type AddSource struct {
	CarID         uint32  `json:"car_id"`
	Price         int     `json:"price"`
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
	ClientPhone   uint64  `json:"client_phone"`
	PhotosCount   uint    `json:"pic_number"`
	Photo         string  `json:"photo"`
	CategoryID    uint16  `json:"category_id"`
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
}

type AppData struct {
	Categories map[uint16]Category
	Currencies map[uint8]Currency
	Models     map[uint16]Model
	Mans       map[uint16]Manufacturer
	Locations  map[uint16]Location
	GearTypes  map[uint16]GearType
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

const url = "https://api2.myauto.ge"
const NumberOfPhotos uint = 5
const MainCategory = 12

var AutoAppData AppData

func ParsePage(page uint16, class string) (uint16, error) {
	loadData()

	addSources, err := loadPage(page, class)
	if err != nil {
		return page, err
	}

	log.Println(class + ": " + strconv.Itoa(len(addSources)) + " Items loaded p " + strconv.Itoa(int(page)))
	if len(addSources) == 0 {
		log.Println(class + ": 0 - resetting page to 1")
		return uint16(1), nil
	} else {
		page++
	}

	carIds := make([]uint32, 0)

	for key := range addSources {
		carIds = append(carIds, key)
	}

	Dbmethods.RestoreTrashedAdds(carIds, class)

	existingAdds, err := Dbmethods.GetExistingAdds(carIds, class)
	log.Print(class+" already exists: ", len(existingAdds), " of ", len(carIds))
	if err != nil {
		log.Println(err)
		return page - 1, err
	}

	var addsToUpdate = make([]Main.Add, 0)
	for id, add := range existingAdds {
		category, err := getCategory(addSources[id])
		if err != nil {
			continue
		}

		add.Name = getName(addSources[id])
		add.Description = getDescription(addSources[id])
		add.Price = addSources[id].Price
		add.Price_usd = addSources[id].PriceUSD
		add.Currency = getCurrency(addSources[id])
		add.Location_id = Dbmethods.GetLocationByAddress(getAddress(addSources[id].LocationId, ""), 0, 0)
		add.CategoryId = category.Id
		add.Images = getImagesUrlList(addSources[id], addSources[id].CarID)

		addsToUpdate = append(addsToUpdate, add)

		delete(addSources, id)
	}

	log.Println(class + " bulk updating " + strconv.Itoa(len(addsToUpdate)))
	Dbmethods.UpdateAddsBulk(addsToUpdate)

	if (len(addSources)) != 0 {
		var addsToInsert = make([]Main.Add, 0)
		for id, addSource := range addSources {
			category, err := getCategory(addSources[id])
			if err != nil {
				continue
			}

			var locationId = Dbmethods.GetLocationByAddress(getAddress(addSource.LocationId, ""), 0, 0)
			add := Main.Add{
				Name:         getName(addSource),
				Description:  getDescription(addSource),
				Price:        addSource.Price,
				Price_usd:    addSource.PriceUSD,
				Currency:     getCurrency(addSource),
				Location_id:  locationId,
				CategoryId:   category.Id,
				Source_class: class,
				Source_id:    id,
				User_id:      getUser(addSource, locationId).Id,
				Images:       getImagesUrlList(addSource, addSource.CarID),
			}

			addsToInsert = append(addsToInsert, add)
		}

		log.Println(class + " bulk going to insert " + strconv.Itoa(len(addsToInsert)))
		Dbmethods.InsertAddsBulk(addsToInsert)
	}

	return page, nil
}

func getImagesUrlList(source AddSource, id uint32) string {
	images := make([]string, 0)
	for i := uint(1); i <= min(source.PhotosCount, NumberOfPhotos); i++ {
		images = append(images, "https://static.my.ge/myauto/photos/"+source.Photo+"/large/"+strconv.Itoa(int(id))+"_"+strconv.Itoa(int(i))+".jpg")
	}

	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

func getUser(addSource AddSource, locationId uint16) Main.User {
	var user, err = Dbmethods.FindUserByPhone(addSource.ClientPhone)
	if err != nil {
		user, err = Dbmethods.CreateUser(addSource.ClientPhone, "ge", getCurrency(addSource), locationId)
	}
	return user
}

func getAddress(locationId uint16, address string) string {
	location, ok := AutoAppData.Locations[locationId]
	if ok {
		address = address + location.Name + ", "
		if location.ParentId != 0 {
			return getAddress(location.ParentId, address)
		}
	} else {
		log.Println(AutoAppData.Locations)
		panic("No location in dictionary " + strconv.Itoa(int(locationId)))
	}

	return address[:len(address)-2]
}

func getCategory(addSource AddSource) (Main.Category, error) {
	manufacturer, manufacturerOk := AutoAppData.Mans[addSource.ManID]
	if !manufacturerOk {
		return Main.Category{}, fmt.Errorf("Manufacturer not found")
	}

	var category Main.Category
	var subCategory Main.Category
	var err = fmt.Errorf("Get category error")

	category, err = Dbmethods.FindCategoryByNameAndParent(manufacturer.Name, MainCategory)
	if err != nil {
		createdCategory, err := Dbmethods.CreateCategory(manufacturer.Name, MainCategory)
		if err != nil {
			return Main.Category{}, err
		}
		category = createdCategory
	}

	subCategoryAuto, subCatOk := AutoAppData.Categories[addSource.CategoryID]
	if !subCatOk {
		return category, nil
	}

	subCategory, err = Dbmethods.FindCategoryByNameAndParent(subCategoryAuto.Name, category.Id)
	if err != nil {
		subCategory, err = Dbmethods.CreateCategory(subCategoryAuto.Name, category.Id)
		if err != nil {
			return category, err
		}
	}

	return subCategory, nil // Return subcategory if found or created successfully
}

func getCurrency(addSource AddSource) string {
	currency, ok := AutoAppData.Currencies[addSource.Currency]
	if ok {
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

	manufacturer, ok := AutoAppData.Mans[addSource.ManID]
	if ok {
		name = append(name, manufacturer.Name)
	}

	model, ok := AutoAppData.Models[addSource.ModelID]
	if ok {
		name = append(name, model.Name)
	}

	if addSource.CarModel != "" {
		name = append(name, addSource.CarModel)
	}

	if addSource.ProdYear != 0 {
		name = append(name, strconv.Itoa(int(addSource.ProdYear)))
	}

	gearType, transmissionOk := AutoAppData.GearTypes[addSource.GearTypeID]

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
			name = append(name, strconv.FormatFloat(float64(float32(addSource.EngineVolume/1000)), 'f', 1, 64)+transmission)
		}

	}

	return strings.Join(name, " ")
}

func loadPage(page uint16, class string) (map[uint32]AddSource, error) {
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

	body, err := Http.LoadUrl(url+"/ka/products/", params)
	if err != nil {
		return nil, err
	}

	var responseObject Response
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}

	result := make(map[uint32]AddSource)

	for _, addSource := range responseObject.Data.AddSources {
		result[addSource.CarID] = addSource
	}

	return result, nil
}

func loadData() {
	once.Do(func() {
		body, err := Http.LoadUrl(url+"/appdata/other_en.json", nil)

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

		currencies := make(map[uint8]Currency)
		for _, currency := range loadedAppData.Currencies {
			currencies[currency.ID] = currency
		}

		AutoAppData = AppData{
			Categories: categories,
			Mans:       mans,
			Locations:  locations,
			Models:     models,
			GearTypes:  gearTypes,
			Currencies: currencies,
		}

		log.Println("Appdata loaded")
	})
}
