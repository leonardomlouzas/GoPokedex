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

type EncounterVersionDetail struct {
	Rate		int				`json:"rate"`
	Version		APIResource		`json:"version"`
}

type EncounterMethodRate struct {
	EncounterMethod		APIResource					`json:"encounter_method"`
	VersionDetails		[]EncounterVersionDetail	`json:"version_details"`
}

type NameEntry struct {
	Language	APIResource		`json:"language"`
	Name		string			`json:"name"`
}

type EncounterDetail struct {
	Chance		int				`json:"chance"`
	Condition	[]APIResource	`json:"condition_values"`
	MaxLevel	int				`json:"max_level"`
	Method		APIResource		`json:"method"`
	MinLevel	int				`json:"min_level"`
}

type PokemonEncounterVersionDetail struct {
	EncounterDetails	[]EncounterDetail	`json:"encounter_details"`
	MaxChance			int					`json:"max_chance"`
	Version				APIResource			`json:"version"`
}

type PokemonEncounter struct {
	Pokemon				APIResource						`json:"pokemon"`
	VersionDetails		[]PokemonEncounterVersionDetail	`json:"version_details"`
}

type LocationAreaDetail struct {
	EncounterMethodRates	[]EncounterMethodRate	`json:"encounter_method_rates"`
	GameIndex				int						`json:"game_index"`
	ID						int						`json:"id"`
	Location				APIResource				`json:"location"`
	Name					string					`json:"name"`
	Names					[]NameEntry				`json:"names"`
	PokemonEncounters		[]PokemonEncounter		`json:"pokemon_encounters"`
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

	var response LocationAreaDetail
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