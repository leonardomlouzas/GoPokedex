package pokeClient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Results struct {
	Name	string	`json:"name"`
	Url		string	`json:"url"`
}

type Response struct {
	Count		int			`json:"count"`
	Next		string		`json:"next"`
	Previous	string		`json:"previous"`
	Results		[]Results	`json:"results"`
}

func GetMap(url string) ([]Results, string, string, error) {
	if url == "" {
		return nil, "", "", fmt.Errorf("url is empty")
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, "", "", fmt.Errorf("error fetching data: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("error: received status code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", "", fmt.Errorf("error reading response body: %v", err)
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, "", "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return response.Results, response.Previous, response.Next, nil
}
