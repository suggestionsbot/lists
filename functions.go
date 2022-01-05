package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joeycumines/go-dotnotation/dotnotation"
	"github.com/pelletier/go-toml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var conn *pgxpool.Pool
var services *toml.Tree

func handleServer() {
	// TODO: Handle authentication
	app := fiber.New(fiber.Config{
		ErrorHandler: formErrorMessage,
	})

	app.Use(recover.New())
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

func fetchStats(httpClient *http.Client, service string, responses *[]BotListServiceResponse, wg *sync.WaitGroup, errors *chan error) {
	token := getServiceToken(service)
	config := getServiceConfig(service)

	req, err := http.NewRequest("GET", config.GetStatsUrl, nil)
	if err != nil {
		*errors <- err
	}

	req.Header.Set("Authorization", token)

	resp, respErr := httpClient.Do(req)
	if respErr != nil {
		*errors <- respErr
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			*errors <- err
		}
	}(resp.Body)

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		*errors <- bodyErr
	}

	var bodyData interface{}
	bodyDataErr := json.Unmarshal(body, &bodyData)
	if bodyDataErr != nil {
		*errors <- bodyDataErr
	}

	var BotListAccessor dotnotation.Accessor

	guildCount, gcErr := BotListAccessor.Get(bodyData, config.Accessor)
	if gcErr != nil {
		*errors <- gcErr
	}

	*responses = append(*responses, BotListServiceResponse{
		Id:         config.Id,
		ShortName:  config.ShortName,
		Url:        config.Url,
		GuildCount: int64(guildCount.(float64)),
	})

	wg.Done()
}

func getServiceConfig(service string) BotListServiceConfig {
	return BotListServiceConfig{
		Id:           services.Get(fmt.Sprintf("services.%s.id", service)).(int64),
		ShortName:    services.Get(fmt.Sprintf("services.%s.short_name", service)).(string),
		LongName:     services.Get(fmt.Sprintf("services.%s.long_name", service)).(string),
		Url:          services.Get(fmt.Sprintf("services.%s.url", service)).(string),
		GetStatsUrl:  services.Get(fmt.Sprintf("services.%s.get_stats_url", service)).(string),
		PostStatsUrl: services.Get(fmt.Sprintf("services.%s.post_stats_url", service)).(string),
		Accessor:     services.Get(fmt.Sprintf("services.%s.accessor", service)).(string),
		Enabled:      services.Get(fmt.Sprintf("services.%s.enabled", service)).(bool),
	}
}

func getServiceToken(service string) string {
	return os.Getenv(fmt.Sprintf("SERVICES_%s_TOKEN", utils.ToUpper(service)))
}
