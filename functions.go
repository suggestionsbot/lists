package main

import (
	bytes2 "bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/keyauth/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joeycumines/go-dotnotation/dotnotation"
	"github.com/pelletier/go-toml"
	"github.com/pelletier/go-toml/query"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var conn *pgxpool.Pool
var config *toml.Tree

func handleServer() {
	app := fiber.New(fiber.Config{
		ErrorHandler: formErrorMessage,
	})

	app.Use(logger.New(logger.Config{
		Format:     config.Get("api.logger.format").(string),
		TimeFormat: config.Get("api.logger.time_format").(string),
		TimeZone:   config.Get("api.logger.timezone").(string),
	}))
	app.Use(keyauth.New(keyauth.Config{
		KeyLookup:    config.Get("api.auth.header_key").(string),
		AuthScheme:   config.Get("api.auth.header_prefix").(string),
		ErrorHandler: formErrorMessage,
		Validator:    validateAuthToken,
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.GetArray("api.cors.allow_origins").(string),
		AllowHeaders: config.GetArray("api.cors.allow_headers").(string),
	}))

	app.Get("/", getRootRoute)

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Post("/guilds", postGuildCountRoute)
	v1.Get("/guilds", getGuildCountRoute)

	v1.Get("/services", getBotListServicesRoute)

	port := os.Getenv("API_PORT")
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
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
	doc, err := toml.LoadFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	config = doc

	fmt.Println("Services config file loaded!")
}

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
	} else if err.Error() != "" {
		message = err.Error()
	}

	return ctx.Status(code).JSON(formJsonBody(
		fiber.Map{
			"code":    code,
			"message": message,
		},
		false,
	))
}

func fetchStats(httpClient *http.Client, config BotListServiceConfig) (*BotListServiceResponse, error) {
	token := getServiceToken(config.ShortName)

	req, err := http.NewRequest("GET", config.GetStatsUrl, nil)
	if err != nil {
		return &BotListServiceResponse{
			ShortName:  config.ShortName,
			Url:        config.Url,
			GuildCount: 0,
			Error:      true,
		}, nil
	}

	req.Header.Set("Authorization", token)

	resp, respErr := httpClient.Do(req)
	if respErr != nil {
		return &BotListServiceResponse{
			ShortName:  config.ShortName,
			Url:        config.Url,
			GuildCount: 0,
			Error:      true,
		}, nil
	}

	defer resp.Body.Close()

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return &BotListServiceResponse{
			ShortName:  config.ShortName,
			Url:        config.Url,
			GuildCount: 0,
			Error:      true,
		}, nil
	}

	var bodyData interface{}
	bodyDataErr := json.Unmarshal(body, &bodyData)
	if bodyDataErr != nil {
		return &BotListServiceResponse{
			ShortName:  config.ShortName,
			Url:        config.Url,
			GuildCount: 0,
			Error:      true,
		}, nil
	}

	var BotListAccessor dotnotation.Accessor
	guildCount, gcErr := BotListAccessor.Get(bodyData, config.Accessor)
	if gcErr != nil {
		return nil, gcErr
	}

	return &BotListServiceResponse{
		ShortName:  config.ShortName,
		Url:        config.Url,
		GuildCount: int64(guildCount.(float64)),
	}, nil
}

func postStatsToBotList(httpClient *http.Client, service BotListServiceConfig, guildCount int64, shardCount int64) error {
	token := getServiceToken(service.ShortName)

	var data = fiber.Map{service.Key: guildCount}
	if service.ShortName == "botsgg" {
		withShardCount := &data
		*withShardCount = fiber.Map{service.Key: guildCount, "shardCount": shardCount}
	}

	jsonData, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		return jsonErr
	}

	req, err := http.NewRequest("POST", service.PostStatsUrl, bytes2.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := httpClient.Do(req)
	if respErr != nil {
		return respErr
	}

	var res fiber.Map
	decErr := json.NewDecoder(resp.Body).Decode(&res)
	if decErr != nil {
		return decErr
	}

	log.Printf("service: %s, code: %d, response: %v", service.ShortName, resp.StatusCode, res)

	return nil
}

func postStatsToBotLists(guildCount int64, shardCount int64) []error {
	wg := sync.WaitGroup{}
	locker := sync.Mutex{}

	var errors []error
	configs := getActiveServices()

	client := &http.Client{Timeout: time.Second * 30}

	for _, config := range configs {
		wg.Add(1)
		go func(c BotListServiceConfig) {
			defer wg.Done()

			err := postStatsToBotList(client, c, guildCount, shardCount)
			if err != nil {
				errors = append(errors, err)
				return
			}

			locker.Lock()
			defer locker.Unlock()

			return
		}(getServiceConfig(config))
	}

	wg.Wait()

	return errors
}

func fetchBotListServiceData() ([]BotListServiceResponse, []error) {
	wg := sync.WaitGroup{}
	locker := sync.Mutex{}

	var responses []BotListServiceResponse
	var errors []error
	configs := getActiveServices()

	client := &http.Client{Timeout: time.Second * 30}

	for _, config := range configs {
		wg.Add(1)
		go func(c BotListServiceConfig) {
			defer wg.Done()

			data, err := fetchStats(client, c)
			if err != nil {
				errors = append(errors, err)
				return
			}

			locker.Lock()
			defer locker.Unlock()

			responses = append(responses, *data)

			return
		}(getServiceConfig(config))
	}

	wg.Wait()

	return responses, errors
}

func getServiceConfig(service string) BotListServiceConfig {
	return BotListServiceConfig{
		ShortName:    config.Get(fmt.Sprintf("services.%s.short_name", service)).(string),
		LongName:     config.Get(fmt.Sprintf("services.%s.long_name", service)).(string),
		Url:          config.Get(fmt.Sprintf("services.%s.url", service)).(string),
		GetStatsUrl:  config.Get(fmt.Sprintf("services.%s.get_stats_url", service)).(string),
		PostStatsUrl: config.Get(fmt.Sprintf("services.%s.post_stats_url", service)).(string),
		Accessor:     config.Get(fmt.Sprintf("services.%s.accessor", service)).(string),
		Key:          config.Get(fmt.Sprintf("services.%s.key", service)).(string),
		Enabled:      config.Get(fmt.Sprintf("services.%s.enabled", service)).(bool),
	}
}

func getVersion() string {
	return config.Get("version").(string)
}

func getServiceToken(service string) string {
	return os.Getenv(fmt.Sprintf("SERVICES_%s_TOKEN", utils.ToUpper(service)))
}

func execQuery(query string, args ...interface{}) (pgconn.CommandTag, error) {
	return conn.Exec(context.Background(), query, args...)
}

func queryRow(query string, args ...interface{}) error {
	return conn.QueryRow(context.Background(), query).Scan(args...)
}

func validateGuildCount(guild GuildCountRequestBody) []*ErrorResponse {
	var errors []*ErrorResponse
	validate := validator.New()
	err := validate.Struct(guild)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}

	return errors
}

func handleBotListErrors(ctx *fiber.Ctx, errors []error) error {
	var data []interface{}
	for _, err := range errors {
		data = append(data, err)
	}

	ctx.Status(fiber.StatusInternalServerError)
	return ctx.JSON(formJsonBody(data, false))
}

func validateAuthToken(_ *fiber.Ctx, token string) (bool, error) {
	tk := os.Getenv("API_TOKEN")

	if token != tk {
		return false, nil
	}

	return true, nil
}

func getActiveServices() []string {
	var services []string

	q, _ := query.Compile("$.services[?(active)].short_name")

	q.SetFilter("active", func(node interface{}) bool {
		if tree, ok := node.(*toml.Tree); ok {
			return tree.Get("enabled").(bool) == true
		}
		return false
	})

	results := q.Execute(config)

	for _, service := range results.Values() {
		services = append(services, service.(string))
	}

	return services
}
