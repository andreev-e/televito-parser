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
	ID    int    `json:"id"`
	Title string `json:"title"`
	Price struct {
		TotalPrice struct {
			Gel int `json:"gel"`
			USD int `json:"usd"`
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

func ParsePage(page uint16) (uint16, error) {
	addSources, err := loadPage(page)
	page++
	if err != nil {
		return page, err
	}

	log.Println(Class + ": " + strconv.Itoa(len(addSources)) + " Items loaded p " + strconv.Itoa(int(page)))
	if len(addSources) == 0 {
		log.Println(Class + ": 0 - resetting page to 1")
		return 0, nil
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
		add.Price = addSources[id].Price.TotalPrice.Gel
		add.Price_usd = float32(addSources[id].Price.TotalPrice.USD)
		add.Currency = "GEL"
		add.Location_id = Dbmethods.GetLocationByAddress(addSources[id].Place, 0, 0)
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

			var locationId = Dbmethods.GetLocationByAddress(addSource.Place, 0, 0)
			user, err := getUser(addSource, locationId)
			if err != nil {
				log.Println(err)
				continue
			}

			add := Main.Add{
				Name:         getName(addSource),
				Description:  getDescription(addSource),
				Price:        addSource.Price.TotalPrice.Gel,
				Price_usd:    float32(addSource.Price.TotalPrice.USD),
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
	images := make([]string, 0)
	for i := 1; i <= min(addSource.Images.Val, numberOfPhotos); i++ {
		images = append(images, "https://static.my.ge/myhome/photos/"+addSource.Images.Path+"/large/"+strconv.Itoa(addSource.ID)+"_"+strconv.Itoa(int(i))+".jpg")
	}

	return "[\"" + strings.Join(images, "\",\"") + "\"]"
}

func getUser(addSource AddSource, locationId uint16) (Main.User, error) {
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

	category, err = Dbmethods.FindCategoryByNameAndParent(addType, mainCategory)
	if err != nil {
		createdCategory, err := Dbmethods.CreateCategory(addType, mainCategory)
		if err != nil {
			return Main.Category{}, err
		}
		category = createdCategory
	}

	subCategoryAuto, subCatOk := estateTypes[addSource.ProductTypeID]
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
	description = append(description, addSource.DescText)

	return strings.Join(description, "\n")
}

func getName(addSource AddSource) string {
	var name []string

	name = append(name, addSource.Title)

	addType, ok := addTypes[addSource.AdtypeID]
	if ok {
		name = append(name, addType)
	}

	estateType, ok := estateTypes[addSource.ProductTypeID]
	if ok {
		name = append(name, estateType)
	}

	return strings.Join(name, " ")
}

func loadPage(page uint16) (map[uint32]AddSource, error) {
	params := map[string]string{
		"Page": strconv.Itoa(int(page)),
	}

	fullUrl := url + "/search?"
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
		return make(map[uint32]AddSource), err
	}

	result := make(map[uint32]AddSource)

	for _, addSource := range responseObject.Data.Children {
		result[uint32(addSource.ID)] = addSource
	}

	return result, nil
}
