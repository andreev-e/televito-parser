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

var categories = map[int]string{
	0:  "13",
	1:  "24",
	2:  "25",
	3:  "62",
	4:  "29",
	5:  "27",
	6:  "33",
	7:  "39",
	8:  "45",
	9:  "42",
	10: "1356",
	11: "48",
	12: "51",
	13: "54",
}

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
	result := make([]Main.Add, 0)

	////fin category by index
	//category, ok := categories[int(page)]
	//if !ok {
	//	return nil, nil
	//}

	minLat, maxLat, minLng, maxLng, err := getLocationBounds(page)

	if err != nil {
		return result, nil
	}

	log.Println("p", page, "minLat: ", minLat, " maxLat: ", maxLat, " minLng: ", minLng, " maxLng: ", maxLng)

	data := map[string]interface{}{
		//"CategoryId": category,
		"SortFields": []map[string]interface{}{
			{
				"FieldName": "ValidFromForDisplay",
				"Ascending": false,
			},
		},
		"GetAllGeolocations": true,
		"ItemsPerPage":       20,
		"PageNumber":         1,
		"fetchBanners":       false,
		"RenderSEOWidget":    false,
		"GeoPolygonQuery": map[string]interface{}{
			"FieldName": "location_rpt",
			"GeoPolygon": []map[string]float64{
				{"Lng": minLng, "Lat": maxLat},
				{"Lng": minLng, "Lat": minLat},
				{"Lng": maxLng, "Lat": minLat},
				{"Lng": maxLng, "Lat": maxLat},
				{"Lng": minLng, "Lat": maxLat},
			},
			"Operation": 2,
		},
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return result, err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Println("error loading " + url)
		return result, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, err
	}

	var responseObject Response
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		log.Printf(string(body))
		return result, err
	}

	log.Println(len(responseObject.AddSources))

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
			add.Price_usd = getPrice(addSource) / 1.08
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

func getLocationBounds(page uint16) (float64, float64, float64, float64, error) {
	const (
		minLatOverall = 41.644183479397455
		maxLatOverall = 46.31658418182218
		minLngOverall = 18.3801373882843
		maxLngOverall = 23.416895338243755
	)

	if page == 0 {
		return minLatOverall, maxLatOverall, minLngOverall, maxLngOverall, nil
	}

	pageCount := uint16(10)

	x := float64(page / pageCount)
	y := float64(page % pageCount)

	xStep := (maxLatOverall - minLatOverall) / float64(pageCount)
	yStep := (maxLngOverall - minLngOverall) / float64(pageCount)

	minLat := minLatOverall + x*xStep
	maxLat := minLatOverall + (x+1)*xStep
	minLng := minLngOverall + y*yStep
	maxLng := minLngOverall + (y+1)*yStep

	if page > 100 {
		return 0, 0, 0, 0, fmt.Errorf("page out of bounds")
	}

	return minLat, maxLat, minLng, maxLng, nil
}

func getImagesUrlList(addSource AddSource) string {
	doc, err := getAddHtml(addSource)
	if err != nil {
		log.Println(err)
		return "[]"
	}

	ulElement := findElementByClass(doc, "span", "pi-img-count-num")
	if ulElement != nil {
		//return "[" + getTextContent(ulElement) + "]"
		return "[]"
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

	ulElement := findElementByClass(doc, "p", "text-description-list product-description short-desc")
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
