package pokeClient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIResource struct {
	Name	string	`json:"name"`
	Url		string	`json:"url"`
}

type MapResponse struct {
	Count		int				`json:"count"`
	Next		string			`json:"next"`
	Previous	string			`json:"previous"`
	Results		[]APIResource	`json:"results"`
}

type minPokemonEncounter struct {
	Pokemon APIResource `json:"pokemon"`
}

type minExploreAreaResponse struct {
	PokemonEncounters []minPokemonEncounter `json:"pokemon_encounters"`
}

type PokemonDetail struct {
	BaseExperience		int		`json:"base_experience"`
	Name				string	`json:"name"`
}

func GetMap(url string) ([]APIResource, string, string, error) {
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

	var response MapResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, "", "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return response.Results, response.Previous, response.Next, nil
}

func GetExploreArea(url string) ([]string, error) {
		if url == "" {
		return nil, fmt.Errorf("url is empty")
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var response minExploreAreaResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Create a map to store unique Pokemon names
	// This will help in avoiding duplicates if/when multiple encounter methods are present
	uniquePokemonsMap := make(map[string]struct{})
	for _, encounter := range response.PokemonEncounters {
		uniquePokemonsMap[encounter.Pokemon.Name] = struct{}{}
	}

	pokemonNames := make([]string, 0, len(uniquePokemonsMap))
	for name := range uniquePokemonsMap {
		pokemonNames = append(pokemonNames, name)
	}

	return pokemonNames, nil
}

func GetPokemonInfo(url string) (PokemonDetail, error) {
		if url == "" {
		return PokemonDetail{}, fmt.Errorf("url is empty")
	}

	res, err := http.Get(url)
	if err != nil {
		return PokemonDetail{}, fmt.Errorf("error fetching data: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return PokemonDetail{}, fmt.Errorf("error: received status code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return PokemonDetail{}, fmt.Errorf("error reading response body: %v", err)
	}

	var pokemonDetails PokemonDetail

	err = json.Unmarshal(body, &pokemonDetails)
	if err != nil {
		return PokemonDetail{}, fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	return pokemonDetails, nil
}