package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/leonardomlouzas/GoPokedex/internal/pokeCache"
	"github.com/leonardomlouzas/GoPokedex/internal/pokeClient"
)

const pokedexURL = "https://pokeapi.co/api/v2/"
const locationAreaURL = pokedexURL + "location-area/"
const pokemonURL = pokedexURL + "pokemon/"

type cliCommands struct {
	name		string
	description string
	callback 	func(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error
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
	pokedex := make(map[string]pokeClient.PokemonDetail)

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
			err := command.callback(&config, cache, commandArg, pokedex) 
			if err != nil {
				fmt.Printf("error executing command '%s': %v\n", commandName, err)
			}
		} else {
			fmt.Printf("unknown command '%s'. Type 'help' for a list of commands.\n", commandName)
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
		"catch": {
			name:			"Catch",
			description:	"Catch a Pokemon.\nUsage: catch <pokemon_name>",
			callback:		commandCatch,
		},
		"inspect": {
			name:			"Inspect",
			description:	"Inspect a Pokemon.\nUsage: inspect <pokemon_name>",
			callback:		commandInspect,
		},
	}
}

func commandExit(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
	fmt.Println("Exiting Pokedex... Bye bye!")
	os.Exit(0)
	return nil
}

func commandHelp(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
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

func commandMap(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
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

func commandMapBack(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
	if conf.prev == "" {
		fmt.Println("you are on the first page")
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

func commandExplore(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
	if arg == "" {
		return fmt.Errorf("an area name must be provided")
	}

	urlToFetch := locationAreaURL + arg
	fmt.Println("Exploring area: ", arg)
	
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
		fmt.Println("no Pokemon found in this area")
	} else {
		fmt.Println("Found pokemon:")
		for _, name := range results {
			fmt.Println("- " + name)
		}
	}

	return nil
}

func commandCatch(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
	if arg == "" {
		return fmt.Errorf("a Pokemon name must be provided")
	}

	pokemonName := strings.ToLower(arg)
	urlToFetch := pokemonURL + pokemonName
	var pokemonData pokeClient.PokemonDetail

	if cachedData, ok := cache.Get(urlToFetch); ok {
		err := json.Unmarshal(cachedData, &pokemonData)
		if err != nil {
			return fmt.Errorf("error unmarshalling cached data: %v", err)
		}
	} else {
		pokemon, err := pokeClient.GetPokemonInfo(urlToFetch)
		if err != nil {
			return fmt.Errorf("error fetching Pokemon data: %v", err)
		}

		pokemonData = pokemon
		dataForCacheBytes, err := json.Marshal(pokemonData)
		if err != nil {
			return fmt.Errorf("error marshalling data for cache: %v", err)
		}

		cache.Add(urlToFetch, dataForCacheBytes)
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// Simulate a random catch attempt based on the Pokemon's base experience.
	attemptValue := rand.Intn(pokemonData.BaseExperience + 1)

	const successThreshold = 40 

	if attemptValue < successThreshold {
		fmt.Printf("%s was caught!\n", pokemonData.Name)
		pokedex[pokemonData.Name] = pokemonData
	} else {
		fmt.Printf("%s escaped!\n", pokemonData.Name)
	}

	return nil
}

func commandInspect(conf *Config, cache *pokeCache.Cache, arg string, pokedex map[string]pokeClient.PokemonDetail) error {
	if arg == "" {
		return fmt.Errorf("a Pokemon name must be provided")
	}
	pokemonName := strings.ToLower(arg)
	if pokemon, ok := pokedex[pokemonName]; ok {
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)
		fmt.Println("Stats:")
		for _, stat := range pokemon.Stats {
			fmt.Printf(" - %s: %d\n", stat.Stat.Name, stat.BaseStat)
		}
		fmt.Println("Types:")
		for _, pokeType := range pokemon.Types {
			fmt.Printf(" - %s\n", pokeType.PokeType.Name)
		}
	} else {
		fmt.Printf("Pokemon %s not found in your Pokedex\n", pokemonName)
	}

	return nil
}