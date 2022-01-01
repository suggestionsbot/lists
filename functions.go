package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var conn *pgxpool.Pool
var services *toml.Tree

func handleServer() {
	// TODO: Handle authentication
	app := fiber.New(fiber.Config{
		ErrorHandler: formErrorMessage,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000, https://api.suggestions.gg, https://suggestions.gg, https://suggestions-voting.ngrok.io",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, User-Agent",
	}))

	app.Get("/", getRootRoute)

	app.Post("/guildCount", postGuildCountRoute)
	app.Get("/guildCount", getGuildCount)

	app.Get("/services", getBotListServices)

	log.Fatal(app.Listen(":3000"))
}

func handleDatabase() {
	dbpool, connErr := pgxpool.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
	if connErr != nil {
		log.Fatal(connErr)
	}

	conn = dbpool

	fmt.Println("PostgreSQL database connected!")
}

func handleServices() {
	doc, err := toml.LoadFile("services.toml")
	if err != nil {
		log.Fatal(err)
	}

	services = doc

	fmt.Println("Services config file loaded!")
}

// TODO: look into if we need to use pointers and such here
func formJsonBody(data interface{}, success bool) fiber.Map {
	return fiber.Map{
		"data":    data,
		"success": success,
		"nonce":   time.Now().UnixMilli(),
	}
}

func formErrorMessage(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "A server side error has occurred."

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return ctx.JSON(formJsonBody(
		fiber.Map{
			"code":    code,
			"message": message,
		},
		false,
	))
}

func getBodyFromBotListService(httpClient *http.Client, service string) (map[string]interface{}, error) {
	url := services.Get(fmt.Sprintf("services.%s.get_stats_url", service)).(string)
	token := os.Getenv(fmt.Sprintf("SERVICES_%s_TOKEN", utils.ToUpper(service)))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)

	resp, respErr := httpClient.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	var bodyData map[string]interface{}
	bodyDataErr := json.Unmarshal([]byte(body), &bodyData)
	if bodyDataErr != nil {
		return nil, bodyDataErr
	}

	return bodyData, nil
}
