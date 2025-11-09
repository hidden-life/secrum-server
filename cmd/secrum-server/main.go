package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("Starting secrum-server...")

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	log.Printf("Environment: %s", env)
}
