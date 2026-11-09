package main

import (
	"time"

	"pokedex/internal/pokecache"
)

func main() {
	cfg := &Config{
		Cache:    pokecache.NewCache(5 * time.Minute),
		Pokedex:  make(map[string]Pokemon),
		Commands: getCommands(),
	}

	startRepl(cfg)
}
