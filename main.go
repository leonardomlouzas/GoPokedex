package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/leonardomlouzas/GoPokedex/internal/pokeCache"
	"github.com/leonardomlouzas/GoPokedex/internal/pokeClient"
)

const pokedexURL = "https://pokeapi.co/api/v2/"
const locationAreaURL = pokedexURL + "location-area/"

type cliCommands struct {
	name		string
	description string
	callback 	func(conf *Config, cache *pokeCache.Cache, arg string) error
}

type Config struct {
	prev	string
	next	string
}

func main() {
	config := Config{
		prev: "",
		next: locationAreaURL,
	}
	reader := bufio.NewScanner(os.Stdin)

	cache := pokeCache.NewCache(5 * time.Second)

	for {
		fmt.Print("Pokedex > ")
		reader.Scan()

		words := cleanInput(reader.Text())
		if len(words) == 0 {
			continue
		}

		var commandArg string 
		commandName := words[0]
		if len(words) > 1 {
			commandArg = words[1]
		}

		if command, ok := getCommands()[commandName]; ok {
			err := command.callback(&config, cache, commandArg) 
			if err != nil {
				fmt.Printf("Error executing command '%s': %v\n", commandName, err)
			}
		} else {
			fmt.Printf("Unknown command '%s'. Type 'help' for a list of commands.\n", commandName)
		}
	}
}

func cleanInput(input string) []string {
	return strings.Fields(strings.ToLower(input))
}

func getCommands() map[string]cliCommands {
	return map[string]cliCommands{
		"exit": {
			name:			"Exit",
			description:	"Exit the Pokedex.",
			callback:		commandExit,
		},
		"help": {
			name:			"Help",
			description:	"List all commands and their descriptions.",
			callback:		commandHelp,
		},
		"map": {
			name:			"Map",
			description:	"Page forward in the Pokedex areas.",
			callback:		commandMap,
		},
		"mapb": {
			name:			"Map Back",
			description:	"Page backward in the Pokedex areas.",
			callback:		commandMapBack,
		},
		"explore": {
			name:			"Explore",
			description:	"Explore a specific area in a map area.\nUsage: explore <area_name>",
			callback:		commandExplore,
		},
	}
}

func commandExit(conf *Config, cache *pokeCache.Cache, arg string) error {
	fmt.Println("Exiting Pokedex... Bye bye!")
	os.Exit(0)
	return nil
}

func commandHelp(conf *Config, cache *pokeCache.Cache, arg string) error {
	fmt.Println("Available commands:")
	commands := getCommands()

	commandNames := make([]string, 0, len(commands))
	for name := range commands {
		commandNames = append(commandNames, name)
	}

	// Sort the command names alphabetically
	sort.Strings(commandNames)

	for _, name := range commandNames {
		fmt.Println("-----> " + commands[name].name)
		fmt.Println(commands[name].description)
	}
	return nil
}

func commandMap(conf *Config, cache *pokeCache.Cache, arg string) error {
	if conf.next == "" {
		fmt.Println("You are on the last page")
		return nil
	}

	urlToFetch := conf.next

	var results []pokeClient.APIResource
	if cachedData, ok := cache.Get(urlToFetch); ok {
		var cachedResponse pokeClient.MapResponse
		err := json.Unmarshal(cachedData, &cachedResponse)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached data: %v", err)
		}

		results = cachedResponse.Results
		conf.prev = cachedResponse.Previous
		conf.next = cachedResponse.Next
	} else {
		apiResults, apiPrev, apiNext, err := pokeClient.GetMap(urlToFetch)
		if err != nil {
			return fmt.Errorf("error fetching map data: %v", err)
		}

		results = apiResults
		conf.prev = apiPrev
		conf.next = apiNext

		responseToCache := pokeClient.MapResponse{
			Results:  apiResults,
			Previous: apiPrev,
			Next:     apiNext,
		}
		dataForCacheBytes, err := json.Marshal(responseToCache)
		if err != nil {
			return fmt.Errorf("error marshalling data for cache: %v", err)
		}
		cache.Add(urlToFetch, dataForCacheBytes)
	}

	for _, result := range results {
		fmt.Printf("Name: %s\n", result.Name)
	}

	return nil
}

func commandMapBack(conf *Config, cache *pokeCache.Cache, arg string) error {
	if conf.prev == "" {
		fmt.Println("You are on the first page")
		return nil
	}

	urlToFetch := conf.prev

	var results []pokeClient.APIResource

	if cachedData, ok := cache.Get(urlToFetch); ok {
		var cachedResponse pokeClient.MapResponse
		err := json.Unmarshal(cachedData, &cachedResponse)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached data: %v", err)
		}
	
		results = cachedResponse.Results
		conf.prev = cachedResponse.Previous
		conf.next = cachedResponse.Next
	} else {
		apiResults, apiPrev, apiNext, err := pokeClient.GetMap(urlToFetch)
		if err != nil {
			return fmt.Errorf("error fetching map data: %v", err)
		}

		results = apiResults
		conf.prev = apiPrev
		conf.next = apiNext

		responseToCache := pokeClient.MapResponse{Results: apiResults, Previous: apiPrev, Next: apiNext}
		dataForCacheBytes, err := json.Marshal(responseToCache)
		if err != nil {
			return fmt.Errorf("error marshalling data for cache: %v", err)
		}

		cache.Add(urlToFetch, dataForCacheBytes)
	}

	for _, result := range results {
		fmt.Printf("Name: %s\n", result.Name)
	}
	return nil
}

func commandExplore(conf *Config, cache *pokeCache.Cache, arg string) error {
	if arg == "" {
		return fmt.Errorf("An area name must be provided.")
	}

	urlToFetch := locationAreaURL + arg
	fmt.Println("Exploring area:", arg)
	
	var results []string

	if cachedData, ok := cache.Get(urlToFetch); ok {
		err := json.Unmarshal(cachedData, &results)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached data: %v", err)
		}
	} else {
		pokemonNames, err := pokeClient.GetExploreArea(urlToFetch)
		if err != nil {
			return fmt.Errorf("error fetching explore area data: %v", err)
		}

		results = pokemonNames
		dataForCacheBytes, err := json.Marshal(results)
		if err != nil {
			return fmt.Errorf("error marshalling data for cache: %v", err)
		}
		cache.Add(urlToFetch, dataForCacheBytes)
	}

	if len(results) == 0 {
		fmt.Println("No Pokemon found in this area.")
	} else {
		fmt.Println("Found pokemon:")
		for _, name := range results {
			fmt.Println("- " + name)
		}
	}

	return nil
}
