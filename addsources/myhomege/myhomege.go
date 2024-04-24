package Myhomege

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"televito-parser/dbmethods"
	Main "televito-parser/models"
)

type Response struct {
	Prs        Prs        `json:"Prs"`
	Cnt        string     `json:"Cnt"`
	Pagination Pagination `json:"Pagination"`
}

type Prs struct {
	Maklers []interface{} `json:"Maklers"`
	Prs     []AddSource   `json:"Prs"`
	Users   Users         `json:"Users"`
}

type AddSource struct {
	ProductID      string              `json:"product_id"`
	UserID         string              `json:"user_id"`
	ParentID       interface{}         `json:"parent_id"`
	MaklerID       interface{}         `json:"makler_id"`
	HasLogo        interface{}         `json:"has_logo"`
	MaklerName     interface{}         `json:"makler_name"`
	LocID          string              `json:"loc_id"`
	StreetAddress  string              `json:"street_address"`
	YardSize       string              `json:"yard_size"`
	YardSizeTypeID string              `json:"yard_size_type_id"`
	AdtypeID       string              `json:"adtype_id"`
	ProductTypeID  string              `json:"product_type_id"`
	Price          string              `json:"price"`
	Photo          string              `json:"photo"`
	PhotoVer       string              `json:"photo_ver"`
	PhotosCount    string              `json:"photos_count"`
	AreaSizeValue  string              `json:"area_size_value"`
	CurrencyID     string              `json:"currency_id"`
	OrderDate      string              `json:"order_date"`
	PriceTypeID    string              `json:"price_type_id"`
	Vip            string              `json:"vip"`
	Color          string              `json:"color"`
	EstateTypeID   string              `json:"estate_type_id"`
	AreaSize       string              `json:"area_size"`
	AreaSizeTypeID string              `json:"area_size_type_id"`
	Comment        string              `json:"comment"`
	MapLat         string              `json:"map_lat"`
	MapLon         string              `json:"map_lon"`
	LLiving        string              `json:"l_living"`
	SpecialPersons string              `json:"special_persons"`
	Rooms          string              `json:"rooms"`
	Bedrooms       string              `json:"bedrooms"`
	Floor          string              `json:"floor"`
	ParkingID      string              `json:"parking_id"`
	Canalization   string              `json:"canalization"`
	Water          string              `json:"water"`
	Road           string              `json:"road"`
	Electricity    string              `json:"electricity"`
	OwnerTypeID    string              `json:"owner_type_id"`
	OsmID          string              `json:"osm_id"`
	NameJSON       string              `json:"name_json"`
	PathwayJSON    string              `json:"pathway_json"`
	Homeselfie     string              `json:"homeselfie"`
	Currencies     map[string]Currency `json:"Currencies"`
	Name           string              `json:"name"`
	Pathway        string              `json:"pathway"`
}

type Currency struct {
	CurrencySymbol string `json:"currency_symbol"`
	CurrencyRate   string `json:"currency_rate"`
	Title          string `json:"title"`
}

type Users struct {
	StatusCode    int                 `json:"StatusCode"`
	StatusMessage string              `json:"StatusMessage"`
	Data          map[string]UserData `json:"Data"`
}

type UserData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type Pagination struct {
	CurrentPage  string `json:"CurrentPage"`
	PerPage      int    `json:"PerPage"`
	ContentCount string `json:"ContentCount"`
}

const (
	Class              = "MyHomeGe"
	url                = "https://api.myhome.ge/ka/"
	numberOfPhotos int = 5
	mainCategory       = 1
)

var userData map[string]UserData

var (
	currencies = map[string]string{
		"1": "USD",
		"2": "EUR",
		"3": "GEL",
	}

	estateTypes = map[string]string{
		"0":  "отель",
		"1":  "квартира",
		"2":  "строительство",
		"3":  "квартира",
		"4":  "торговое",
		"5":  "магазин",
		"6":  "подвал",
		"7":  "производство",
		"8":  "склад",
		"9":  "коммерческая(?9)",
		"10": "коммерческая",
		"12": "гараж",
		"13": "участок сельхоз",
		"14": "участок",
		"15": "участок коммерческий",
		"17": "новостройка",
		"18": "вилла",
		"21": "участок под застройку",
	}

	addTypes = map[string]string{
		"1": "Продажа",
		"2": "Залог",
		"3": "Аренда",
		"7": "Посуточно",
		"8": "Аренда",
	}
)

func ParsePage(page uint16) (uint16, error) {
	addSources, err := loadPage(page)
	if err != nil {
		return page, err
	}

	log.Println(Class + ": " + strconv.Itoa(len(addSources)) + " Items loaded p " + strconv.Itoa(int(page)))
	if len(addSources) == 0 {
		log.Println(Class + ": 0 - resetting page to 1")
		return uint16(1), nil
	} else {
		page++
	}

	carIds := make([]uint32, 0)

	for key := range addSources {
		carIds = append(carIds, key)
	}

	Dbmethods.RestoreTrashedAdds(carIds, Class)

	existingAdds, err := Dbmethods.GetExistingAdds(carIds, Class)
	log.Print(Class+" already exists: ", len(existingAdds), " of ", len(carIds))
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

		currency, ok := currencies[addSources[id].CurrencyID]
		if !ok {
			currency = "USD"
		}

		price, _ := strconv.ParseInt(addSources[id].Price, 10, 32)
		add.Name = getName(addSources[id])
		add.Description = getDescription(addSources[id])
		add.Price = int(price)
		add.Price_usd = float32(price)
		add.Currency = currency
		add.Location_id = Dbmethods.GetLocationByAddress(addSources[id].StreetAddress, 0, 0)
		add.CategoryId = category.Id
		add.Images = getImagesUrlList(addSources[id])

		addsToUpdate = append(addsToUpdate, add)

		delete(addSources, id)
	}

	Dbmethods.UpdateAddsBulk(addsToUpdate)

	if (len(addSources)) != 0 {
		var addsToInsert = make([]Main.Add, 0)
		for id, addSource := range addSources {
			category, err := getCategory(addSources[id])
			if err != nil {
				log.Println(err)
				continue
			}

			var locationId = Dbmethods.GetLocationByAddress(addSource.StreetAddress, 0, 0)
			user, err := getUser(addSource, locationId)
			if err != nil {
				log.Println(err)
				continue
			}

			currency, ok := currencies[addSources[id].CurrencyID]
			if !ok {
				currency = "USD"
			}

			price, _ := strconv.ParseInt(addSource.Price, 10, 32)

			add := Main.Add{
				Name:         getName(addSource),
				Description:  getDescription(addSource),
				Price:        int(price),
				Price_usd:    float32(price),
				Currency:     currency,
				Location_id:  locationId,
				CategoryId:   category.Id,
				Source_class: Class,
				Source_id:    id,
				User_id:      user.Id,
				Images:       getImagesUrlList(addSource),
			}

			addsToInsert = append(addsToInsert, add)
		}

		Dbmethods.InsertAddsBulk(addsToInsert)
	}

	return page, nil
}

func getImagesUrlList(addSource AddSource) string {
	images := make([]string, 0)
	photosCount, _ := strconv.Atoi(addSource.PhotosCount)
	for i := 1; i <= min(photosCount, numberOfPhotos); i++ {
		images = append(images, "https://static.my.ge/myhome/photos/"+addSource.Photo+"/large/"+addSource.ProductID+"_"+strconv.Itoa(int(i))+".jpg")
	}

	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

func getUser(addSource AddSource, locationId uint16) (Main.User, error) {
	currency, ok := currencies[addSource.CurrencyID]
	if !ok {
		currency = "USD"
	}

	var user, err = Dbmethods.FindUserByPhone(userData[addSource.UserID].Username)
	if err != nil {
		log.Println(err)
		user, err = Dbmethods.CreateUser(userData[addSource.UserID].Username, "ge", currency, locationId, nil)
	}

	return user, nil
}

func getCategory(addSource AddSource) (Main.Category, error) {
	addType, addTypeOk := addTypes[addSource.AdtypeID]
	if !addTypeOk {
		return Main.Category{}, fmt.Errorf("Manufacturer not found")
	}

	var category Main.Category
	var subCategory Main.Category
	var err = fmt.Errorf("Get category error")

	category, err = Dbmethods.FindCategoryByNameAndParent(addType, mainCategory)
	if err != nil {
		createdCategory, err := Dbmethods.CreateCategory(addType, mainCategory)
		if err != nil {
			return Main.Category{}, err
		}
		category = createdCategory
	}

	subCategoryAuto, subCatOk := estateTypes[addSource.EstateTypeID]
	if !subCatOk {
		return category, nil
	}

	subCategory, err = Dbmethods.FindCategoryByNameAndParent(addType+" "+subCategoryAuto, category.Id)
	if err != nil {
		subCategory, err = Dbmethods.CreateCategory(addType+" "+subCategoryAuto, category.Id)
		if err != nil {
			return category, err
		}
	}

	return subCategory, nil
}

func getDescription(addSource AddSource) string {
	var description []string
	description = append(description, addSource.Comment)

	if addSource.Rooms != "" && addSource.Rooms != "0" {
		description = append(description, "Комнат: "+addSource.Rooms)
	}

	if addSource.Bedrooms != "" && addSource.Bedrooms != "0" {
		description = append(description, "Спален: "+addSource.Bedrooms)
	}

	if addSource.YardSize != "" && addSource.YardSize != "0" {
		description = append(description, "Площадь двора: "+addSource.YardSize)
	}

	if addSource.Water == "1" {
		description = append(description, "Вода")
	}

	if addSource.Road == "1" {
		description = append(description, "Дорога")
	}

	if addSource.Electricity == "1" {
		description = append(description, "Электричество")
	}

	if addSource.Canalization == "1" {
		description = append(description, "Канализация")
	}

	return strings.Join(description, "\n")
}

func getName(addSource AddSource) string {
	var name []string

	addType, ok := addTypes[addSource.AdtypeID]
	if ok {
		name = append(name, addType)
	}

	estateType, ok := estateTypes[addSource.EstateTypeID]
	if ok {
		name = append(name, estateType)
	}

	if addSource.Rooms != "" && addSource.Rooms != "0" {
		name = append(name, addSource.Rooms+"к")
	}

	name = append(name, addSource.AreaSizeValue+"m²")

	if addSource.Floor != "" {
		name = append(name, "эт."+addSource.Floor)
	}

	return strings.Join(name, " ")
}

func loadPage(page uint16) (map[uint32]AddSource, error) {
	params := map[string]string{
		"Page":      strconv.Itoa(int(page)),
		"Ajax":      "1",
		"WithPhoto": "1",
		"WithMap":   "1",
	}

	fullUrl := url + "/products/?"
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

	result := make(map[uint32]AddSource)

	userData = responseObject.Prs.Users.Data

	for _, addSource := range responseObject.Prs.Prs {
		id, err := strconv.ParseUint(addSource.ProductID, 10, 32)
		if err == nil {
			result[uint32(id)] = addSource
		}
	}

	return result, nil
}
