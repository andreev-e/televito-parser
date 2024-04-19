package Ssge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"televito-parser/dbmethods"
	Main "televito-parser/models"
)

type AddSource struct {
	ApplicationId uint32 `json:"applicationId"`
	Address       struct {
		CityTitle string `json:"cityTitle"`
	} `json:"address"`
	Price struct {
		PriceGeo     uint32 `json:"priceGeo"`
		UnitPriceGeo uint8  `json:"unitPriceGeo"`
		PriceUSD     uint32 `json:"priceUsd"`
		UnitPriceUSD uint8  `json:"unitPriceUsd"`
		CurrencyType uint8  `json:"currencyType"`
	} `json:"price"`
	AppImages []struct {
		FileName  string `json:"fileName"`
		IsMain    bool   `json:"isMain"`
		Is360     bool   `json:"is360"`
		OrderNo   uint8  `json:"orderNo"`
		ImageType uint8  `json:"imageType"`
	} `json:"appImages"`
	ImageCount           uint8       `json:"imageCount"`
	Title                string      `json:"title"`
	ShortTitle           string      `json:"shortTitle"`
	Description          string      `json:"description"`
	TotalArea            float32     `json:"totalArea"`
	TotalAmountOfFloor   float32     `json:"totalAmountOfFloor"`
	FloorNumber          string      `json:"floorNumber"`
	NumberOfBedrooms     uint8       `json:"numberOfBedrooms"`
	Type                 uint8       `json:"type"`
	DealType             uint8       `json:"dealType"`
	IsMovedUp            bool        `json:"isMovedUp"`
	IsHighlighted        bool        `json:"isHighlighted"`
	IsUrgent             bool        `json:"isUrgent"`
	VipStatus            uint8       `json:"vipStatus"`
	HasRemoteViewing     bool        `json:"hasRemoteViewing"`
	VideoLink            string      `json:"videoLink"`
	CommercialRealEstate uint8       `json:"commercialRealEstateType"`
	OrderDate            string      `json:"orderDate"`
	CreateDate           string      `json:"createDate"`
	UserId               string      `json:"userId"`
	IsFavorite           bool        `json:"isFavorite"`
	IsForUkraine         bool        `json:"isForUkraine"`
	IsHidden             bool        `json:"isHidden"`
	IsUserHidden         bool        `json:"isUserHidden"`
	IsConfirmed          bool        `json:"isConfirmed"`
	DetailUrl            string      `json:"detailUrl"`
	HomeId               interface{} `json:"homeId"`
	UserInfo             struct {
		Name     string `json:"name"`
		Image    string `json:"image"`
		UserType uint8  `json:"userType"`
	} `json:"userInfo"`
	SimilarityGroup interface{} `json:"similarityGroup"`
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
		add.Price_usd = float32(addSources[id].Price.PriceUSD)
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
				Price_usd:    float32(addSource.Price.PriceUSD),
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
	userInformation struct {
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

	body, err := ioutil.ReadAll(response.Body)

	var responseObject UserResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return user, err
	}

	var userPhone = ""
	for _, phone := range responseObject.userInformation.Phones {
		userPhone = phone
		break
	}

	user, err = Dbmethods.FindUserByPhone(userPhone)

	if err != nil {
		user, err = Dbmethods.CreateUser(userPhone, "ge", "GEL", locationId)
	}
	return user, nil

	//        $userData = json_decode($response->getBody()->getContents(), false, 512, JSON_THROW_ON_ERROR);
	//
	//        $location = $this->getLocation($addSource);
	//
	//        $user = User::query()->firstOrCreate([
	//            'contact' => $userData->userInformatino->phones[0],
	//        ], [
	//            'source_id' => $addSource->userId,
	//            'lang' => Languages::ge->name,
	//            'currency' => Currencies::gel,
	//            'location_id' => $location->id,
	//        ]);
	//
	//        if ($user instanceof User) {
	//            return $user;
	//        }
	//
	//        throw new RuntimeException('User not found');
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
		"pageSize": strconv.Itoa(pageSize), // Assuming pageSize is defined elsewhere
		"order":    "1",
		"page":     strconv.Itoa(int(page)), // Assuming page is defined elsewhere
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

	body, err := ioutil.ReadAll(response.Body)

	var responseObject Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
		log.Printf(response.Header.Get("www-authenticate"))
		return nil, err
	}

	result := make(map[uint32]AddSource)

	for _, addSource := range responseObject.AddSources {
		result[addSource.ApplicationId] = addSource
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

		body, err := ioutil.ReadAll(response.Body)
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
