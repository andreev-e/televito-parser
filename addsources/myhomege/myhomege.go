package Myhomege

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	Dbmethods "televito-parser/dbmethods"
	Main "televito-parser/models"
	"time"
)

type Response struct {
	Data struct {
		Title       string      `json:"title"`
		Type        string      `json:"type"`
		Children    []AddSource `json:"children"`
		CurrentPage int         `json:"current_page"`
		LastPage    int         `json:"last_page"`
		PerPage     int         `json:"per_page"`
		Total       int         `json:"total"`
		From        int         `json:"from"`
		To          int         `json:"to"`
	} `json:"data"`
	Success bool `json:"success"`
}

type AddSource struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
	Price struct {
		TotalPrice struct {
			Gel float32 `json:"gel"`
			USD float32 `json:"usd"`
		} `json:"total_price"`
		SQMPrice struct {
			Gel float64 `json:"gel"`
			USD float64 `json:"usd"`
		} `json:"sqm_price"`
	} `json:"price"`
	ProductTypeID int    `json:"product_type_id"`
	AdtypeID      int    `json:"adtype_id"`
	DescText      string `json:"desc_text"`
	Place         string `json:"place"`
	Date          int    `json:"date"`
	OwnerTypeID   int    `json:"owner_type_id"`
	UserID        int    `json:"user_id"`
	Images        struct {
		Val      int    `json:"val"`
		PhotoVer int    `json:"photo_ver"`
		Path     string `json:"path"`
	} `json:"images"`
}

const (
	Class              = "MyHomeGe"
	url                = "https://api2.myhome.ge/api/ru/"
	numberOfPhotos int = 5
	mainCategory       = 1
)

var (
	estateTypes = map[int]string{
		0:  "отель",
		1:  "квартира",
		2:  "строительство",
		3:  "квартира",
		4:  "торговое",
		5:  "магазин",
		6:  "подвал",
		7:  "производство",
		8:  "склад",
		9:  "коммерческая(?9)",
		10: "коммерческая",
		12: "гараж",
		13: "участок сельхоз",
		14: "участок",
		15: "участок коммерческий",
		17: "новостройка",
		18: "вилла",
		21: "участок под застройку",
	}

	addTypes = map[int]string{
		1: "Продажа",
		2: "Залог",
		3: "Аренда",
		7: "Посуточно",
		8: "Аренда",
	}
)

func LoadPage(page uint16, class string) ([]Main.Add, error) {
	params := map[string]string{
		"Page": strconv.Itoa(int(page)),
	}

	fullUrl := url + "search?"
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
		log.Printf(string(body))
		return make([]Main.Add, 0), err
	}

	result := make([]Main.Add, 0)

	for _, addSource := range responseObject.Data.Children {
		locationId := Dbmethods.GetLocationIdByAddress(addSource.Place, 0, 0)
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
			add.Price = addSource.Price.TotalPrice.Gel
			add.Price_usd = addSource.Price.TotalPrice.USD
			add.Source_class = class
			add.Source_id = addSource.ID
			add.CategoryId = category.ID
			add.Images = getImagesUrlList(addSource)
			add.Currency = "GEL"
			add.UpdatedAt = time.Now()

			result = append(result, add)
		}

	}
	return result, nil
}

func getImagesUrlList(addSource AddSource) string {
	images := make([]string, 0)
	for i := 1; i <= min(addSource.Images.Val, numberOfPhotos); i++ {
		url := "https://static.my.ge/myhome/photos/" + addSource.Images.Path + "/large/" + strconv.Itoa(int(addSource.ID)) + "_" + strconv.Itoa(int(i)) + ".jpg"
		images = append(images, strings.ReplaceAll(url, "/", "\\/"))
	}

	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

func getUser(addSource AddSource, locationId uint64) (Main.User, error) {
	userName := strconv.Itoa(addSource.UserID)
	var user, err = Dbmethods.FindUserByPhone(userName)
	if err != nil {
		user, err = Dbmethods.CreateUser(userName, "ge", "GEL", locationId, nil)
	}

	return user, err
}

func getCategory(addSource AddSource) (Main.Category, error) {
	addType, addTypeOk := addTypes[addSource.AdtypeID]
	if !addTypeOk {
		return Main.Category{}, fmt.Errorf("Manufacturer not found")
	}

	var category Main.Category
	var subCategory Main.Category
	var err = fmt.Errorf("Get category error")

	createdCategory, err := Dbmethods.RetrieveCategory(addType, mainCategory)
	if err != nil {
		return Main.Category{}, err
	}
	category = createdCategory

	subCategoryAuto, subCatOk := estateTypes[addSource.ProductTypeID]
	if !subCatOk {
		return category, nil
	}

	subCategory, err = Dbmethods.RetrieveCategory(addType+" "+subCategoryAuto, category.ID)
	if err != nil {
		return category, err
	}

	return subCategory, nil
}

func getDescription(addSource AddSource) string {
	var description []string
	description = append(description, addSource.DescText)

	return strings.Join(description, "\n")
}

func getName(addSource AddSource) string {
	var name []string

	name = append(name, addSource.Title)

	addType, ok := addTypes[addSource.AdtypeID]
	if ok {
		name = append(name, addType)
	} else {
		estateType, ok := estateTypes[addSource.ProductTypeID]
		if ok {
			name = append(name, estateType)
		}
	}

	return strings.Join(name, " ")
}
