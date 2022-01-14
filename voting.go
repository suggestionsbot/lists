package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	if err := godotenv.Load(".env.local", ".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal("Error loading .env file")
	}

	handleDatabase()
	handleServices()
}

func main() {
	fmt.Println("Voting v1.0 - Copyright (c) 2021 Anthony Collier")

	handleServer()
}
