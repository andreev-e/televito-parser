package Http

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func LoadUrl(url string, params map[string]string) []byte {
	fullUrl := url + "?"
	for key, value := range params {
		fullUrl += key + "=" + value + "&"
	}

	response, err := http.Get(fullUrl)

	if err != nil {
		log.Printf("Error fetching data: %v\n", err)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body: %v\n", err)
		return nil
	}

	return body
}
