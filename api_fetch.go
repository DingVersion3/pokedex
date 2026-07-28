package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// getOrFetch returns the raw bytes for url, using the cache when possible
// and populating it on a real fetch.
func getOrFetch(url string, cfg *Config) ([]byte, error) {
	if cfg.Cache != nil {
		if data, ok := cfg.Cache.Get(url); ok {
			fmt.Println("cache hit!")
			return data, nil
		}
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("bad response fetching %s: status %d: %s", url, res.StatusCode, body)
	}

	if cfg.Cache != nil {
		cfg.Cache.Add(url, body)
	}

	return body, nil
}

func fetchLocations(url string, cfg *Config) ([]string, error) {
	body, err := getOrFetch(url, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}

	var apiRes PokeAPIResponse
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	cfg.Next = apiRes.Next
	cfg.Previous = apiRes.Previous

	var names []string
	for _, area := range apiRes.Results {
		names = append(names, area.Name)
	}

	return names, nil
}

func fetchLocationArea(areaName string, cfg *Config) ([]string, error) {
	url := "https://pokeapi.co/api/v2/location-area/" + areaName

	body, err := getOrFetch(url, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch location area: %w", err)
	}

	var area LocationArea
	if err := json.Unmarshal(body, &area); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	var pokemonNames []string
	for _, encounter := range area.PokemonEncounters {
		pokemonNames = append(pokemonNames, encounter.Pokemon.Name)
	}

	return pokemonNames, nil
}

func fetchPokemon(name string, cfg *Config) (Pokemon, error) {
	url := "https://pokeapi.co/api/v2/pokemon/" + name

	body, err := getOrFetch(url, cfg)
	if err != nil {
		return Pokemon{}, fmt.Errorf("failed to fetch pokemon: %w", err)
	}

	var pokemon Pokemon
	if err := json.Unmarshal(body, &pokemon); err != nil {
		return Pokemon{}, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return pokemon, nil
}
