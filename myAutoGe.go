package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type AddSource struct {
	CarID    int32   `json:"car_id"`
	Price    int     `json:"price"`
	PriceUSD float32 `json:"price_usd"`
	Currency int8    `json:"currency_id"`
}

type Response struct {
	Data struct {
		AddSources []AddSource `json:"items"`
	} `json:"data"`
}

const url = "https://api2.myauto.ge/ka/products/"

func MyAutoGeParsePage(page int32) int32 {
	addSources := loadPage(page)

	if len(addSources) == 0 {
		fmt.Println("0 Items - resetting page to 1")
	} else {
		page++
	}

	carIds := make([]int32, 0)
	for _, addSource := range addSources {
		carIds = append(carIds, addSource.CarID)
	}

	existingAdds := GetExistingAdds(carIds)

	fmt.Println(existingAdds)

	return page
}

func loadPage(page int32) []AddSource {
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

	return responseObject.Data.AddSources
}
