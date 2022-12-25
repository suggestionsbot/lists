package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func init() {
	if err := godotenv.Load(".env.local", ".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal("Error loading .env file")
	}

	loadDatabase()
	loadConfig()
}

func main() {
	version := getVersion()
	date := time.Now()
	year := date.Year()

	message := fmt.Sprintf("Lists %s - Copyright (c) %d Anthony Collier", version, year)
	fmt.Println(message)

	handleServer()
}
