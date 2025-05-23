package main

import (
	"testing"
	"time"

	"github.com/leonardomlouzas/GoPokedex/internal/pokeCache"
	"github.com/leonardomlouzas/GoPokedex/internal/pokeClient"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  ",
			expected: []string{},
		},
		{
			input:    "  hello  ",
			expected: []string{"hello"},
		},
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  HellO  World  ",
			expected: []string{"hello", "world"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("lengths don't match: '%v' vs '%v'", actual, c.expected)
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%v) == %v, expected %v", c.input, actual, c.expected)
			}
		}
	}
}

func TestCommandPokedex(t *testing.T) {
	cache := pokeCache.NewCache(5 * time.Second)
	config := Config{}

	pokedex := make(map[string]pokeClient.PokemonDetail)
	pokemonName := "pikachu"
	pokemonDetail := pokeClient.PokemonDetail{
		BaseExperience:	100,
		Name:           "pikachu",
		Height:         4,
		Weight:         60,
		Stats: []pokeClient.PokeStat{},
		Types: []pokeClient.PokemonType{},
	}
	
	// Expect the Pokedex to be empty
	commandPokedex(&config, cache, pokemonName, pokedex)
	if len(pokedex) != 0 {
		t.Errorf("expected empty pokedex, got %v", pokedex)
	}

	// Add a Pokemon to the Pokedex
	pokedex[pokemonName] = pokemonDetail
	commandPokedex(&config, cache, pokemonName, pokedex)
	if len(pokedex) != 1 {
		t.Errorf("expected Pokedex with 1 Pokemon, got %v", pokedex)
	}

	// Check if the Pokemon is in the Pokedex
	if _, ok := pokedex[pokemonName]; !ok {
		t.Errorf("expected %s to be in pokedex, got %v", pokemonName, pokedex)
	}
}