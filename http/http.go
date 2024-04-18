package Http

import (
	"io"
	"io/ioutil"
	"net/http"
)

func LoadUrl(url string, params map[string]string) ([]byte, error) {
	fullUrl := url + "?"
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

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
