package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"net/http"
	"strconv"
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

func main() {

	page := 1
	url := "https://api2.myauto.ge/ka/products/?ForRent=0&Page=" + strconv.Itoa(page)

	// Make an HTTP GET request
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching data: %v\n", err)
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// Parse JSON data
	var responseObject Response
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	fmt.Println("Items:")

	//////

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
