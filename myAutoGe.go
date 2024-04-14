package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	GearTypes  []GearType     `json:"GearTypes:items"`
	Locations  struct {
		Items []MyAutoGeLocation `json:"items"`
	} `json:"Locations"`
}

type AppData struct {
	Categories map[uint16]Category
	Currencies map[uint8]Currency
	Models     map[uint16]Model
	Mans       map[uint16]Manufacturer
	Locations  map[uint16]MyAutoGeLocation
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

type MyAutoGeLocation struct {
	ID       uint16 `json:"location_id"`
	Name     string `json:"title"`
	ParentId uint16 `json:"parent_id"`
}

type GearType struct {
	ID   uint16 `json:"gear_type_id"`
	Name string `json:"title"`
}

const url = "https://api2.myauto.ge"
const sourceClass = "MyAutoGe"

var appData AppData

func MyAutoGeParsePage(page uint16) uint16 {
	loadData()

	addSources := loadPage(page)

	if len(addSources) == 0 {
		fmt.Println("0 Items - resetting page to 1")
	} else {
		fmt.Println(strconv.Itoa(len(addSources)) + " Items loaded")
		page++
	}

	carIds := make([]uint32, 0)

	for key, _ := range addSources {
		carIds = append(carIds, key)
	}

	RestoreTrashedAdds(carIds, sourceClass)

	existingAdds := GetExistingAdds(carIds, sourceClass)

	for id, add := range existingAdds {
		add.name = getName(addSources[id])
		add.description = getDescription(addSources[id])
		add.price = addSources[id].Price
		add.price_usd = addSources[id].PriceUSD
		add.currency = getCurrency(addSources[id])
		add.location_id = getLocationByAddress(getAddress(addSources[id].LocationId, ""), 0, 0)
		add.categoryId = getCategory(addSources[id])

		UpdateAdd(add)

		delete(addSources, id)
	}

	fmt.Println(strconv.Itoa(len(addSources)) + " Items adding")

	for id, addSource := range addSources {
		var locationId = getLocationByAddress(getAddress(addSource.LocationId, ""), 0, 0)
		add := Add{
			name:         getName(addSource),
			description:  getDescription(addSource),
			price:        addSource.Price,
			price_usd:    addSource.PriceUSD,
			currency:     getCurrency(addSource),
			location_id:  locationId,
			categoryId:   getCategory(addSource),
			source_class: sourceClass,
			source_id:    id,
			user_id:      getUser(addSource, locationId).id,
		}

		InsertAdd(add)

		fmt.Println("inserted add " + strconv.Itoa(int(id)))
	}

	return page
}

func getUser(addSource AddSource, locationId uint16) User {
	var user, err = findUserByPhone(addSource.ClientPhone)
	if err != nil {
		user, err = createNewUser(addSource.ClientPhone, "ge", getCurrency(addSource), locationId)
	}
	return user
}

func getAddress(locationId uint16, address string) string {
	location, ok := appData.Locations[locationId]
	if ok {
		address = address + location.Name + ", "
		if location.ParentId != 0 {
			return getAddress(location.ParentId, address)
		}
	} else {
		fmt.Println(appData.Locations)
		panic("No location in dictionary " + strconv.Itoa(int(locationId)))
	}

	return address[:len(address)-2]
}

func getCategory(addSource AddSource) int {
	return 0
}

func getCurrency(addSource AddSource) string {
	currency, ok := appData.Currencies[addSource.Currency]
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

	manufacturer, ok := appData.Mans[addSource.ManID]
	if ok {
		name = append(name, manufacturer.Name)
	}

	model, ok := appData.Models[addSource.ModelID]
	if ok {
		name = append(name, model.Name)
	}

	if addSource.CarModel != "" {
		name = append(name, addSource.CarModel)
	}

	if addSource.ProdYear != 0 {
		name = append(name, strconv.Itoa(int(addSource.ProdYear)))
	}

	gearType, transmissionOk := appData.GearTypes[addSource.GearTypeID]

	if ok && addSource.EngineVolume != 0 {
		var transmission = ""

		if transmissionOk {
			switch gearType.Name {
			case "Automatic":
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
			name = append(name, strconv.Itoa(int(addSource.EngineVolume/1000))+transmission)
		}

	}

	return strings.Join(name, " ")
}

func loadPage(page uint16) map[uint32]AddSource {
	params := map[string]string{
		"ForRent":       "0",
		"CurrencyID":    "1",
		"MileageType":   "1",
		"SortOrder":     "1",
		"Page":          strconv.Itoa(int(page)),
		"hideDealPrice": "1",
		"Locs":          "2.3.4.7.15.30.113.52.37.36.38.39.40.31.5.41.44.47.48.53.54.8.16.6.14.13.12.11.10.9.55.56.57.59.58.61.62.63.64.66.71.72.74.75.76.77.78.80.81.82.83.84.85.86.87.88.91.96.97.101.109",
	}

	body := LoadUrl(url+"/ka/products/", params)

	var responseObject Response
	err := json.Unmarshal(body, &responseObject)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil
	}

	result := make(map[uint32]AddSource)

	for _, addSource := range responseObject.Data.AddSources {
		result[addSource.CarID] = addSource
	}

	return result
}

func loadData() {
	if len(appData.Categories) == 0 {
		body := LoadUrl(url+"/appdata/other_en.json", nil)

		var loadedAppData LoadedAppData

		err := json.Unmarshal(body, &loadedAppData)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
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

		locations := make(map[uint16]MyAutoGeLocation)
		for _, location := range loadedAppData.Locations.Items {
			locations[location.ID] = location
		}

		models := make(map[uint16]Model)
		for _, model := range loadedAppData.Models {
			models[model.ID] = model
		}

		gearTypes := make(map[uint16]GearType)
		for _, gearType := range loadedAppData.GearTypes {
			gearTypes[gearType.ID] = gearType
		}

		currencies := make(map[uint8]Currency)
		for _, currency := range loadedAppData.Currencies {
			currencies[currency.ID] = currency
		}

		appData = AppData{
			Categories: categories,
			Mans:       mans,
			Locations:  locations,
			Models:     models,
			GearTypes:  gearTypes,
			Currencies: currencies,
		}

		fmt.Println("Appdata loaded")
	}
}
