package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"pokedex/internal/pokecache"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	cfg := &Config{
		Cache:   pokecache.NewCache(5 * time.Second),
		Pokedex: make(map[string]Pokemon),
	}

	commands := getCommands()

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()

		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		cmd, ok := commands[commandName]
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
