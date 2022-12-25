package main

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/suggestionsbot/lists/docs"
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

//	@title			Suggestions Lists
//	@version		1.1
//	@description	This is the API documentation for the Lists API, responsible for intetracting with the various bot lists that Suggestions is listed on.
//	@termsOfService	https://suggestions.gg/terms

//	@contact.name	Suggestions
//	@contact.url	https://suggestions.bot/discord
//	@contact.email	hello@suggestions.gg

//	@license.name	AGPL-3.0
//	@license.url	https://github.com/suggestionsbot/lists/blob/main/LICENSE

//	@tag.name			General
//	@tag.description	All routes for the service.

// @securityDefinitions	APIKeyHeader
// @in						header
//
// @name					Authorization
// @description			The API key used to secure all API routes, preventing unauthorized access.
func main() {
	version := getVersion()
	date := time.Now()
	year := date.Year()

	message := fmt.Sprintf("Lists %s - Copyright (c) %d Anthony Collier", version, year)
	fmt.Println(message)

	handleServer()
}
