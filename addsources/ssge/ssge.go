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
	Dbmethods "televito-parser/dbmethods"
	Main "televito-parser/models"
)

type UserResponse struct {
	UserInformation struct {
		Phones []string `json:"phones"`
	} `json:"userInformatino"`
}

type Response struct {
	AddSources []AddSource `json:"realStateItemModel"`
}

type AddSource struct {
	ApplicationID            uint64      `json:"applicationId"`
	Status                   int         `json:"status"`
	Address                  Address     `json:"address"`
	Price                    Price       `json:"price"`
	AppImages                []AppImage  `json:"appImages"`
	ImageCount               int         `json:"imageCount"`
	Title                    string      `json:"title"`
	ShortTitle               string      `json:"shortTitle"`
	Description              string      `json:"description"`
	TotalArea                float64     `json:"totalArea"`
	TotalAmountOfFloor       float64     `json:"totalAmountOfFloor"`
	FloorNumber              string      `json:"floorNumber"`
	NumberOfBedrooms         int         `json:"numberOfBedrooms"`
	Type                     int         `json:"type"`
	DealType                 int         `json:"dealType"`
	IsMovedUp                bool        `json:"isMovedUp"`
	IsHighlighted            bool        `json:"isHighlighted"`
	IsUrgent                 bool        `json:"isUrgent"`
	VIPStatus                int         `json:"vipStatus"`
	HasRemoteViewing         bool        `json:"hasRemoteViewing"`
	VideoLink                interface{} `json:"videoLink"`
	CommercialRealEstateType int         `json:"commercialRealEstateType"`
	OrderDate                string      `json:"orderDate"`
	CreateDate               string      `json:"createDate"`
	UserID                   string      `json:"userId"`
	HomeID                   interface{} `json:"homeId"`
	UserInfo                 interface{} `json:"userInfo"`
	SimilarityGroup          interface{} `json:"similarityGroup"`
}

type Address struct {
	CityTitle string `json:"cityTitle"`
}

type Price struct {
	PriceGeo float32 `json:"priceGeo"`
	PriceUSD float32 `json:"priceUsd"`
}

type AppImage struct {
	FileName string `json:"fileName"`
}

const (
	Class              = "SSGe"
	url                = "https://api-gateway.ss.ge/v1/RealEstate/"
	numberOfPhotos int = 5
	mainCategory       = 1
	pageSize           = 30
)

var (
	token    = ""
	addTypes = map[int]string{
		1: "Аренда",
		2: "Залог",
		3: "Посуточно",
		4: "Продажа",
	}
	estateTypes = map[int]string{
		1: "дача",
		2: "гостиница",
		3: "участок",
		4: "дом",
		5: "квартира",
		6: "коммерческая",
	}
)

func LoadPage(page uint16, class string) ([]Main.Add, error) {
	var token, err = getToken()
	if err != nil {
		token = ""
		return nil, err
	}

	requestBodyData := map[string]interface{}{
		"page":     page,
		"pageSize": pageSize,
		"order":    1,
	}

	requestBodyJSON, err := json.Marshal(requestBodyData)
	if err != nil {
		token = ""
		return nil, err
	}

	requestBody := bytes.NewReader(requestBodyJSON)

	headers := map[string]string{
		"Accept-Language": "ru",
		"Authorization":   "Bearer " + token,
		"Content-Type":    "application/json",
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url+"LegendSearch", requestBody)
	if err != nil {
		token = ""
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := client.Do(req)
	if err != nil {
		token = ""
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		token = ""
		return nil, err
	}

	var responseObject Response
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		log.Printf("Error parsing page JSON: %v\n", err)
		log.Printf("Raw response body: %s\n", string(body))
		token = ""

		return nil, err
	}

	result := make([]Main.Add, 0)

	for _, addSource := range responseObject.AddSources {
		address := getAddress(addSource)
		locationId := Dbmethods.GetLocationIdByAddress(address, 0, 0)
		user, userError := getUser(addSource, locationId)
		category, categoryError := getCategory(addSource)

		if userError == nil && categoryError == nil {
			var add Main.Add
			add.User_id = user.Id
			add.Status = 2
			add.Approved = 1
			add.Location_id = locationId
			add.Name = getName(addSource)
			add.Description = getDescription(addSource)
			add.Price = addSource.Price.PriceGeo
			add.Price_usd = addSource.Price.PriceUSD
			add.Source_class = class
			add.Source_id = addSource.ApplicationID
			add.CategoryId = category.Id
			add.Images = getImagesUrlList(addSource)
			add.Currency = "GEL"

			result = append(result, add)
		}

	}
	return result, nil
}

func getImagesUrlList(addSource AddSource) string {
	var images = make([]string, 0)
	for index, image := range addSource.AppImages {
		if index >= numberOfPhotos {
			break
		}
		images = append(images, strings.ReplaceAll(image.FileName, "/", "\\/"))

	}
	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

func getUser(addSource AddSource, locationId uint64) (Main.User, error) {
	var user Main.User
	var err error
	user, err = Dbmethods.FindUserBySourceId(addSource.UserID)
	if err == nil {
		return user, nil
	}

	token, err = getToken()
	if err != nil {
		return user, err
	}

	requestBodyData := map[string]interface{}{
		"page":     1,
		"pageSize": 1,
		"userId":   addSource.UserID,
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
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		log.Printf("Error parsing user %v JSON: %v\n", addSource.UserID, err)
		log.Printf(string(body))
		log.Println(token)
		return user, err
	}

	var userPhone = ""
	for _, phone := range responseObject.UserInformation.Phones {
		userPhone = phone
		break
	}

	user, err = Dbmethods.FindUserByPhone(userPhone)
	if err != nil {
		user, err = Dbmethods.CreateUser(userPhone, "ge", "GEL", locationId, addSource.UserID)
	}

	return user, err
}

func getAddress(addSource AddSource) string {
	return addSource.Address.CityTitle + ", Georgia"
}

func getCategory(addSource AddSource) (Main.Category, error) {
	var categoryName = getCategoryName(addSource)
	return Dbmethods.RetrieveCategory(categoryName, mainCategory)
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
		if estateTypes[addSource.Type] != "" {
			name = append(name, estateTypes[addSource.Type])
		} else {
			name = append(name, "type:"+strconv.Itoa(int(addSource.Type)))
		}
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
