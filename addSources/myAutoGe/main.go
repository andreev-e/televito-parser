package myAutoGe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"televito-parser/helpers"
)

type Response struct {
	Data struct {
		Items []struct {
			CarID    int     `json:"car_id"`
			Price    int     `json:"price"`
			PriceUSD float32 `json:"price_usd"`
			Currency int8    `json:"currency_id"`
		} `json:"items"`
	} `json:"data"`
}

const url = "https://api2.myauto.ge/ka/products/"

func ParsePage(page int) {
	params := map[string]string{
		"ForRent":       "0",
		"CurrencyID":    "1",
		"MileageType":   "1",
		"SortOrder":     "1",
		"Page":          strconv.Itoa(page),
		"hideDealPrice": "1",
		"Locs":          "2.3.4.7.15.30.113.52.37.36.38.39.40.31.5.41.44.47.48.53.54.8.16.6.14.13.12.11.10.9.55.56.57.59.58.61.62.63.64.66.71.72.74.75.76.77.78.80.81.82.83.84.85.86.87.88.91.96.97.101.109",
	}
	body := helpers.LoadUrl(url, params)
	// Make an HTTP GET request

	// Parse JSON data
	var responseObject Response
	err := json.Unmarshal(body, &responseObject)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	fmt.Println("Items:")
	for _, item := range responseObject.Data.Items {
		fmt.Println(item)
	}

	db, err := sql.Open("mysql", "televito_and:BeD8Pf00ZBxGqvGr@tcp(159.253.19.143:3306)/televito_and")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Execute a query
	rows, err := db.Query("SELECT id, user_id, name FROM adds LIMIT 0, 10")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Process the rows
	for rows.Next() {
		var id int
		var user_id int
		var name string

		err := rows.Scan(&id, &user_id, &name)
		if err != nil {
			panic(err)
		}

		fmt.Println(id, user_id, name)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		panic(err)
	}
}
