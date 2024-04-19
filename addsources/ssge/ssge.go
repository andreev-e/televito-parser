package Ssge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"televito-parser/dbmethods"
	Main "televito-parser/models"
)

type Address struct {
	MunicipalityId    interface{} `json:"municipalityId"`
	MunicipalityTitle interface{} `json:"municipalityTitle"`
	CityId            int         `json:"cityId"`
	CityTitle         string      `json:"cityTitle"`
	DistrictId        int         `json:"districtId"`
	DistrictTitle     string      `json:"districtTitle"`
	SubdistrictId     int         `json:"subdistrictId"`
	SubdistrictTitle  string      `json:"subdistrictTitle"`
	StreetId          int         `json:"streetId"`
	StreetTitle       string      `json:"streetTitle"`
	StreetNumber      string      `json:"streetNumber"`
}

type Price struct {
	PriceGeo     int `json:"priceGeo"`
	UnitPriceGeo int `json:"unitPriceGeo"`
	PriceUsd     int `json:"priceUsd"`
	UnitPriceUsd int `json:"unitPriceUsd"`
	CurrencyType int `json:"currencyType"`
}

type Image struct {
	FileName  string `json:"fileName"`
	IsMain    bool   `json:"isMain"`
	Is360     bool   `json:"is360"`
	OrderNo   int    `json:"orderNo"`
	ImageType int    `json:"imageType"`
}

type UserInfo struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	UserType int    `json:"userType"`
}

type AddSource struct {
	ApplicationId            int         `json:"applicationId"`
	Status                   int         `json:"status"`
	Address                  Address     `json:"address"`
	Price                    Price       `json:"price"`
	AppImages                []Image     `json:"appImages"`
	ImageCount               int         `json:"imageCount"`
	Title                    string      `json:"title"`
	ShortTitle               string      `json:"shortTitle"`
	Description              string      `json:"description"`
	TotalArea                float64     `json:"totalArea"`
	TotalAmountOfFloor       float64     `json:"totalAmountOfFloor"`
	FloorNumber              string      `json:"floorNumber"`
	NumberOfBedrooms         int         `json:"numberOfBedrooms"`
	Type                     uint8       `json:"type"`
	DealType                 uint8       `json:"dealType"`
	IsMovedUp                bool        `json:"isMovedUp"`
	IsHighlighted            bool        `json:"isHighlighted"`
	IsUrgent                 bool        `json:"isUrgent"`
	VipStatus                int         `json:"vipStatus"`
	HasRemoteViewing         bool        `json:"hasRemoteViewing"`
	VideoLink                interface{} `json:"videoLink"`
	CommercialRealEstateType int         `json:"commercialRealEstateType"`
	OrderDate                string      `json:"orderDate"`
	CreateDate               string      `json:"createDate"`
	UserId                   string      `json:"userId"`
	IsFavorite               bool        `json:"isFavorite"`
	IsForUkraine             bool        `json:"isForUkraine"`
	IsHidden                 bool        `json:"isHidden"`
	IsUserHidden             bool        `json:"isUserHidden"`
	IsConfirmed              bool        `json:"isConfirmed"`
	DetailUrl                string      `json:"detailUrl"`
	HomeId                   interface{} `json:"homeId"`
	UserInfo                 UserInfo    `json:"userInfo"`
	SimilarityGroup          interface{} `json:"similarityGroup"`
}

type Response struct {
	AddSources []AddSource `json:"realStateItemModel"`
}

const (
	Class              = "SSGe"
	url                = "https://api-gateway.ss.ge/v1/RealEstate/"
	numberOfPhotos int = 5
	mainCategory       = 1
	pageSize           = 16
)

var (
	token    = ""
	addTypes = map[uint8]string{
		uint8(1): "Аренда",
		uint8(2): "Залог",
		uint8(3): "Посуточно",
		uint8(4): "Продажа",
	}
	estateTypes = map[uint8]string{
		1: "дача",
		2: "гостиница",
		3: "участок",
		4: "дом",
		5: "квартира",
		6: "коммерческая",
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

		add.Name = getName(addSources[id])
		add.Description = getDescription(addSources[id])
		add.Price = int(addSources[id].Price.PriceGeo)
		add.Price_usd = float32(addSources[id].Price.PriceUsd)
		add.Currency = "GEL"
		add.Location_id = Dbmethods.GetLocationByAddress(getAddress(addSources[id]), 0, 0)
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
				continue
			}

			var locationId = Dbmethods.GetLocationByAddress(getAddress(addSource), 0, 0)
			user, err := getUser(addSource, locationId)
			if err != nil {
				continue
			}

			add := Main.Add{
				Name:         getName(addSource),
				Description:  getDescription(addSource),
				Price:        int(addSource.Price.PriceGeo),
				Price_usd:    float32(addSource.Price.PriceUsd),
				Currency:     "GEL",
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
	var images = make([]string, 0)
	for index, image := range addSource.AppImages {
		if index >= numberOfPhotos {
			break
		}
		images = append(images, image.FileName)

	}
	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

type UserResponse struct {
	UserInformation struct {
		Phones []string `json:"phones"`
	} `json:"userInformatino"`
}

func getUser(addSource AddSource, locationId uint16) (Main.User, error) {
	var user Main.User
	var err error
	user, err = Dbmethods.FindUserBySourceId(addSource.UserId)
	if err == nil {
		return user, nil
	}

	token, err = getToken()
	if err != nil {
		return user, err
	}

	requestBodyData := map[string]interface{}{
		"pageSize": "0", // Assuming pageSize is defined elsewhere
		"userId":   addSource.UserId,
		"page":     "1",
	}

	requestBodyJSON, err := json.Marshal(requestBodyData)
	if err != nil {
		return user, err
	}

	requestBody := bytes.NewReader(requestBodyJSON)

	headers := map[string]string{
		"Accept-Language": "ru",
		"Authorization":   "Bearer " + token,
		"Content-Type":    "application/json",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537." + strconv.Itoa(rand.Intn(40)),
		"OS":              "web",
		"Referer":         "https://home.ss.ge/",
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url+"user-all-applications", requestBody)
	if err != nil {
		return user, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := client.Do(req)
	if err != nil {
		return user, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	var responseObject UserResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		return user, err
	}

	var userPhone = ""
	for _, phone := range responseObject.UserInformation.Phones {
		userPhone = phone
		break
	}

	user, err = Dbmethods.FindUserByPhone(userPhone)

	return Dbmethods.CreateUser(userPhone, "ge", "GEL", locationId)
}

func getAddress(addSource AddSource) string {
	return addSource.Address.CityTitle + ", Georgia"
}

func getCategory(addSource AddSource) (Main.Category, error) {
	var categoryName = getCategoryName(addSource)
	category, err := Dbmethods.FindCategoryByNameAndParent(categoryName, mainCategory)
	if err == nil {
		return category, nil
	}
	return Dbmethods.CreateCategory(categoryName, mainCategory)
}

func getCategoryName(addSource AddSource) string {
	var result = []string{}
	result = append(result, addTypes[addSource.DealType])

	if estateTypes[addSource.Type] != "" {
		result = append(result, estateTypes[addSource.Type])
	}

	return strings.Join(result, " ")
}

func getDescription(addSource AddSource) string {
	var description = []string{}
	description = append(description, addSource.Description)

	if addSource.NumberOfBedrooms != 0 {
		description = append(description, "Спален "+strconv.Itoa(int(addSource.NumberOfBedrooms)))
	}

	return strings.Join(description, "\n")
}

func getName(addSource AddSource) string {
	var name = []string{}
	if addTypes[addSource.DealType] != "" {
		name = append(name, addTypes[addSource.DealType])
	} else {
		name = append(name, "dealType:"+strconv.Itoa(int(addSource.DealType)))
	}

	if estateTypes[addSource.Type] != "" {
		name = append(name, estateTypes[addSource.Type])
	} else {
		name = append(name, "type:"+strconv.Itoa(int(addSource.Type)))
	}

	name = append(name, strconv.Itoa(int(addSource.TotalArea))+"m²")

	if addSource.FloorNumber != "" {
		var floor = "эт." + addSource.FloorNumber

		if addSource.TotalAmountOfFloor != 0 {
			floor += "(" + strconv.Itoa(int(addSource.TotalAmountOfFloor)) + ")"
		}

		name = append(name, floor)
	}

	return strings.Join(name, " ")
}

func loadPage(page uint16) (map[uint32]AddSource, error) {
	var token, err = getToken()
	if err != nil {
		return nil, err
	}

	requestBodyData := map[string]interface{}{
		"pageSize": pageSize,
		"order":    1,
		"page":     page,
	}

	requestBodyJSON, err := json.Marshal(requestBodyData)
	if err != nil {
		return nil, err
	}

	requestBody := bytes.NewReader(requestBodyJSON)

	headers := map[string]string{
		"Accept-Language": "ru",
		"Authorization":   "Bearer " + token,
		"Content-Type":    "application/json",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537." + strconv.Itoa(rand.Intn(40)),
		"OS":              "web",
		"Referer":         "https://home.ss.ge/",
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url+"LegendSearch", requestBody)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var responseObject Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		log.Printf("Raw response body: %s\n", body)
		return nil, err
	}

	log.Println(body)
	log.Println(responseObject)
	log.Println(token)

	result := make(map[uint32]AddSource)

	for _, addSource := range responseObject.AddSources {
		result[uint32(addSource.ApplicationId)] = addSource
	}

	return result, nil
}

func getToken() (string, error) {
	if token == "" {
		response, err := http.Get("https://home.ss.ge/ka/udzravi-qoneba")
		if err != nil {
			return "", err
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", err
		}

		re := regexp.MustCompile(`"credentialsToken":"(.*?)"`)
		match := re.FindStringSubmatch(string(body))
		if len(match) > 1 {
			token = match[1]
		} else {
			return "", fmt.Errorf("unable to find token")
		}
	}
	return token, nil
}
