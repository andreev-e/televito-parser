package Halooglasi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	Dbmethods "televito-parser/dbmethods"
	Main "televito-parser/models"
	"time"
)

const (
	url                = "https://www.halooglasi.com/Quiddita.Widgets.Ad/AdCategoryBasicSearchWidgetAux/GetSidebarData"
	numberOfPhotos int = 10
	mainCategory       = 1
)

type AddSource struct {
	ID       string `json:"Id"`
	Title    string `json:"Title"`
	ListHTML string `json:"ListHTML"`
}

type Response struct {
	AdGeoLocations []struct {
		Coord string `json:"Coord"`
		Id    uint64 `json:"Id"`
	} `json:"AdGeoLocations"`
	AddSources []AddSource `json:"Ads"`
	PageNumber int         `json:"PageNumber"`
	TotalPages int         `json:"TotalPages"`
}

func LoadPage(page uint16, class string) ([]Main.Add, error) {
	data := map[string]interface{}{
		"CategoryId": "24",
		"SortFields": []map[string]interface{}{
			{
				"FieldName": "ValidFromForDisplay",
				"Ascending": false,
			},
		},
		"GetAllGeolocations": true,
		"ItemsPerPage":       20,
		"PageNumber":         page,
		"fetchBanners":       false,
		"BaseTaxonomy":       "/nekretnine/prodaja-kuca",
		"RenderSEOWidget":    false,
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Println("error loading " + url)
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

	for _, addSource := range responseObject.AddSources {
		addressString, addressError := getAddress(addSource)

		lat, lng := getLatLng(responseObject, addSource.ID)

		locationId := Dbmethods.GetLocationIdByAddress(addressString, lat, lng)
		user, userError := getUser(addSource, locationId)

		id, _ := strconv.ParseInt(addSource.ID, 10, 64)

		if userError == nil && addressError == nil {
			var add Main.Add
			add.UserId = user.ID
			add.Status = 2
			add.Approved = 1
			add.Location_id = locationId
			add.Name = addSource.Title
			add.Description = getDescription(addSource)
			add.Price = getPrice(addSource)
			add.Price_usd = 0
			add.Source_class = class
			add.Source_id = uint64(id)
			add.CategoryId = mainCategory
			add.Images = getImagesUrlList(addSource)
			add.Currency = "EUR"
			add.UpdatedAt = time.Now()

			//add.Characteristics = getCharacteristics(addSource)

			result = append(result, add)
		}

	}

	return result, nil
}

func getImagesUrlList(addSource AddSource) string {
	doc, err := getAddHtml(addSource)
	if err != nil {
		log.Println(err)
		return "[]"
	}

	ulElement := findElementByClass(doc, "span", "pi-img-count-num")
	if ulElement != nil {
		return "[" + getTextContent(ulElement) + "]"
	}

	return "[]"
}

func getPrice(addSource AddSource) float32 {
	doc, err := getAddHtml(addSource)
	if err != nil {
		log.Println(err)
		return 0
	}

	ulElement := findElementByClass(doc, "div", "central-feature")
	if ulElement != nil {
		priceString := ulElement.FirstChild.Attr[0].Val
		priceString = strings.ReplaceAll(priceString, ".", "")
		price, err := strconv.ParseFloat(priceString, 32)
		if err != nil {
			log.Println(err)
			return 0
		}
		log.Println(price)
		return float32(price)
	}

	return 0
}

func getAddHtml(source AddSource) (*html.Node, error) {
	decodedHtml := html.UnescapeString(source.ListHTML)
	//log.Println(decodedHtml)
	return html.Parse(strings.NewReader(decodedHtml))
}

func getDescription(source AddSource) string {
	doc, err := getAddHtml(source)
	if err != nil {
		log.Println(err)
		return ""
	}

	ulElement := findElementByClass(doc, "p", "short-desc")
	if ulElement != nil {
		return getTextContent(ulElement)
	}

	return ""
}

func getUser(addSource AddSource, locationId uint64) (Main.User, error) {
	userName := addSource.ID
	var user, err = Dbmethods.FindUserByPhone(userName)
	if err != nil {
		user, err = Dbmethods.CreateUser(userName, "rs", "RSD", locationId, nil)
	}

	return user, err
}

func getLatLng(responseObject Response, adGeoLocationId string) (float32, float32) {
	for _, adGeoLocation := range responseObject.AdGeoLocations {
		if strconv.FormatUint(adGeoLocation.Id, 10) == adGeoLocationId {
			coords := strings.Split(adGeoLocation.Coord, ",")
			lat, err := strconv.ParseFloat(coords[1], 32)
			if err != nil {
				return 0, 0
			}
			lng, err := strconv.ParseFloat(coords[0], 32)
			if err != nil {
				return 0, 0
			}
			return float32(lat), float32(lng)
		}
	}
	return 0, 0
}

func getAddress(source AddSource) (string, error) {
	doc, err := getAddHtml(source)
	if err != nil {
		log.Println(err)
		return "", err
	}

	ulElement := findElementByClass(doc, "ul", "subtitle-places")
	if ulElement != nil {
		items := extractListItems(ulElement)
		joined := strings.Join(items, ", ")
		return joined, nil
	}

	return "", fmt.Errorf("could not find address in HTML")
}

func findElementByClass(n *html.Node, tag, class string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == class {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findElementByClass(c, tag, class); result != nil {
			return result
		}
	}
	return nil
}

func extractListItems(ul *html.Node) []string {
	var items []string
	for li := ul.FirstChild; li != nil; li = li.NextSibling {
		if li.Type == html.ElementNode && li.Data == "li" {
			items = append(items, getTextContent(li))
		}
	}
	return items
}

func getTextContent(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}

	return strings.TrimSpace(sb.String())
}
