package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	pokecache "github.com/leonardomlouzas/GoPokedex/internal/pokeCache"
	pokeclient "github.com/leonardomlouzas/GoPokedex/internal/pokeClient"
)

const pokedexURL = "https://pokeapi.co/api/v2/"

type cliCommands struct {
	name		string
	description string
	callback 	func(conf *Config, cache *pokecache.Cache) error
}

type Config struct {
	prev	string
	next	string
}

func main() {
	config := Config{
		prev: "",
		next: pokedexURL + "location-area/",
	}
	reader := bufio.NewScanner(os.Stdin)

	cache := pokecache.NewCache(5 * time.Second)

	for {
		fmt.Print("Pokedex > ")
		reader.Scan()

		words := cleanInput(reader.Text())
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		if command, ok := getCommands()[commandName]; ok {
			err := command.callback(&config, cache)
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
	}
}

func commandExit(conf *Config, cache *pokecache.Cache) error {
	fmt.Println("Exiting Pokedex... Bye bye!")
	os.Exit(0)
	return nil
}

func commandHelp(conf *Config, cache *pokecache.Cache) error {
	fmt.Println("Available commands:")
	for _, command := range getCommands() {
		fmt.Println("-----> " + command.name)
		fmt.Println(command.description)
	}
	return nil
}

func commandMap(conf *Config, cache *pokecache.Cache) error {
	if conf.next == "" {
		fmt.Println("You are on the last page")
		return nil
	}

	urlToFetch := conf.next

	var results []pokeclient.Results
	if cachedData, ok := cache.Get(urlToFetch); ok {
		var cachedResponse pokeclient.Response
		err := json.Unmarshal(cachedData, &cachedResponse)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached data: %v", err)
		}

		results = cachedResponse.Results
		conf.prev = cachedResponse.Previous
		conf.next = cachedResponse.Next
	} else {
		apiResults, apiPrev, apiNext, err := pokeclient.GetMap(urlToFetch)
		if err != nil {
			return fmt.Errorf("error fetching map data: %v", err)
		}

		results = apiResults
		conf.prev = apiPrev
		conf.next = apiNext

		responseToCache := pokeclient.Response{
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

func commandMapBack(conf *Config, cache *pokecache.Cache) error {
	if conf.prev == "" {
		fmt.Println("You are on the first page")
		return nil
	}

	urlToFetch := conf.prev

	var results []pokeclient.Results

	if cachedData, ok := cache.Get(urlToFetch); ok {
		var cachedResponse pokeclient.Response
		err := json.Unmarshal(cachedData, &cachedResponse)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached data: %v", err)
		}
	
		results = cachedResponse.Results
		conf.prev = cachedResponse.Previous
		conf.next = cachedResponse.Next
	} else {
		apiResults, apiPrev, apiNext, err := pokeclient.GetMap(urlToFetch)
		if err != nil {
			return fmt.Errorf("error fetching map data: %v", err)
		}

		results = apiResults
		conf.prev = apiPrev
		conf.next = apiNext

		responseToCache := pokeclient.Response{Results: apiResults, Previous: apiPrev, Next: apiNext}
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