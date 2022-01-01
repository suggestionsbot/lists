package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}

	handleDatabase()
	handleServices()
}

func main() {
	fmt.Println("Voting v1.0 - Copyright (c) 2021 Anthony Collier")

	handleServer()
}
