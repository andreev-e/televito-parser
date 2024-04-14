package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type AddSource struct {
	CarID        uint32  `json:"car_id"`
	Price        int     `json:"price"`
	PriceUSD     float32 `json:"price_usd"`
	Currency     uint8   `json:"currency_id"`
	ManID        uint16  `json:"man_id"`
	ModelID      uint16  `json:"model_id"`
	CarModel     string  `json:"car_model"`
	ProdYear     uint16  `json:"prod_year"`
	EngineVolume uint16  `json:"engine_volume"`
	GearTypeID   uint16  `json:"gear_type_id"`
	VehicleType  uint16  `json:"vehicle_type"`
}

type Response struct {
	Data struct {
		AddSources []AddSource `json:"items"`
	} `json:"data"`
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
	ID   uint16 `json:"location_id"`
	Name string `json:"title"`
}

type GearType struct {
	ID   uint16 `json:"gear_type_id"`
	Name string `json:"title"`
}

type LoadedAppData struct {
	Categories []Category
	Currencies []Currency
	Models     []Model
	Mans       []Manufacturer
	Locations  []Location
	GearTypes  []GearType
}

type AppData struct {
	Categories map[uint16]Category
	Currencies map[uint16]Currency
	Models     map[uint16]Model
	Mans       map[uint16]Manufacturer
	Locations  map[uint16]Location
	GearTypes  map[uint16]GearType
}

const url = "https://api2.myauto.ge/ka/products/"

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

	RestoreTrashedAdds(carIds, "MyAutoGe")

	existingAdds := GetExistingAdds(carIds, "MyAutoGe")

	for id, add := range existingAdds {
		add.name = getName(addSources[id])
		//            'name' => $name,
		//            'description' => $this->getAddDescription($addSource),
		//            'price' => $addSource->price,
		//            'price_usd' => $addSource->price_usd,
		//            'currency' => $currency,
		//            'location_id' => $location->id,
		//            'category_id' => $category->id,
	}

	fmt.Println(existingAdds)

	return page
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

	body := LoadUrl(url, params)

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

		locations := make(map[uint16]Location)
		for _, location := range loadedAppData.Locations {
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

		appData = AppData{
			Categories: categories,
			Mans:       mans,
			Locations:  locations,
			Models:     models,
			GearTypes:  gearTypes,
		}

		fmt.Println("Appdata loaded")
	}
}
