package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"

	"github.com/sudeeya/key-exchange/internal/agent"
)

func main() {
	envFile := flag.String("e", ".env", "Path to the file storing environment variables")

	flag.Parse()

	if err := godotenv.Load(*envFile); err != nil {
		log.Fatal(err)
	}

	a := agent.NewAgent()
	a.Run()
}
