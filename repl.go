package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
)

// catchThreshold controls catch difficulty. A Pokemon's base experience
// is used as the upper bound of a random draw; the draw must land at or
// below catchThreshold for the catch to succeed. Pokemon with higher base
// experience have a wider range to draw from, so they're harder to catch.
const catchThreshold = 40

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *Config, params []string) error
}

func startRepl(cfg *Config) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()

		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		cmd, ok := cfg.Commands[commandName]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		err := cmd.callback(cfg, words[1:])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Lists the Pokemon found in a given location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Throws a Pokeball at a Pokemon, adding it to your Pokedex if caught",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Prints details about a caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Lists the names of all Pokemon you've caught",
			callback:    commandPokedex,
		},
	}
}

func commandExit(cfg *Config, params []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, params []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range cfg.Commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func cleanInput(text string) []string {
	lowered := strings.ToLower(text)
	fields := strings.Fields(lowered)
	return fields
}

func commandMap(cfg *Config, params []string) error {
	url := "https://pokeapi.co/api/v2/location-area"
	if cfg.Next != nil {
		url = *cfg.Next
	}

	locations, err := fetchLocations(url, cfg)
	if err != nil {
		return err
	}

	for _, name := range locations {
		fmt.Println(name)
	}
	return nil
}

func commandMapb(cfg *Config, params []string) error {
	if cfg.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	locations, err := fetchLocations(*cfg.Previous, cfg)
	if err != nil {
		return err
	}

	for _, name := range locations {
		fmt.Println(name)
	}
	return nil
}

func commandExplore(cfg *Config, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("you must provide a location area name, e.g. explore pastoria-city-area")
	}
	areaName := params[0]

	fmt.Printf("Exploring %s...\n", areaName)

	pokemonNames, err := fetchLocationArea(areaName, cfg)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, name := range pokemonNames {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}

func commandCatch(cfg *Config, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("you must provide a pokemon name, e.g. catch pikachu")
	}
	pokemonName := params[0]

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	pokemon, err := fetchPokemon(pokemonName, cfg)
	if err != nil {
		return err
	}

	baseExperience := pokemon.BaseExperience
	if baseExperience < 1 {
		baseExperience = 1
	}

	throwResult := rand.Intn(baseExperience)
	if throwResult > catchThreshold {
		fmt.Printf("%s escaped!\n", pokemonName)
		return nil
	}

	if cfg.Pokedex == nil {
		cfg.Pokedex = make(map[string]Pokemon)
	}
	cfg.Pokedex[pokemonName] = pokemon
	fmt.Printf("%s was caught!\n", pokemonName)
	return nil
}

func commandInspect(cfg *Config, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("you must provide a pokemon name, e.g. inspect pidgey")
	}
	pokemonName := params[0]

	pokemon, ok := cfg.Pokedex[pokemonName]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *Config, params []string) error {
	if len(cfg.Pokedex) == 0 {
		fmt.Println("you haven't caught any pokemon yet")
		return nil
	}

	names := make([]string, 0, len(cfg.Pokedex))
	for name := range cfg.Pokedex {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Your Pokedex:")
	for _, name := range names {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}
