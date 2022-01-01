package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"sync"
	"time"
)

func getRootRoute(ctx *fiber.Ctx) error {
	return ctx.JSON(formJsonBody(
		fiber.Map{"message": "Hello world!"},
		true,
	))
}

// TODO: Handle validation
func postGuildCountRoute(ctx *fiber.Ctx) error {
	var jsonObj GuildCountResponse
	if err := json.Unmarshal(ctx.Body(), &jsonObj); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	commandTags, execErr := conn.Exec(context.Background(), fmt.Sprintf("insert into guildcount(guild_count, timestamp) values(%d, %d) returning *", jsonObj.Count, jsonObj.Timestamp))
	if execErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, execErr.Error())
	}

	fmt.Printf("Rows affected: %d", commandTags.RowsAffected())

	return ctx.JSON(formJsonBody(GuildCountResponse{
		Count: jsonObj.Count,
	}, true))
}

func getGuildCount(ctx *fiber.Ctx) error {
	var guildCount int64
	var timestamp int64
	err := conn.QueryRow(context.Background(), "select guild_count, timestamp from guildcount order by timestamp desc fetch first row only").Scan(&guildCount, &timestamp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(formJsonBody(
		GuildCountResponse{
			Count:     guildCount,
			Timestamp: timestamp,
		}, true,
	))
}

func getBotListServices(ctx *fiber.Ctx) error {
	//var topgg *BotListServiceResponse
	//var botsgg *BotListServiceResponse
	//var dlspace *BotListServiceResponse
	//var dbl *BotListServiceResponse
	//var discords *BotListServiceResponse

	var responses []*BotListServiceResponse

	var wg sync.WaitGroup
	wg.Add(1)

	client := &http.Client{Timeout: time.Second * 30}

	// top.gg
	go func() {
		body, reqErr := getBodyFromBotListService(client, "topgg")
		if reqErr != nil {
			fiber.NewError(fiber.StatusInternalServerError, reqErr.Error())
		}

		responses = append(responses, &BotListServiceResponse{
			Id:         services.Get("services.topgg.id").(int64),
			ShortName:  services.Get("services.topgg.short_name").(string),
			GuildCount: int64(body["server_count"].(float64)),
			Enabled:    services.Get("services.topgg.enabled").(bool),
		})

		wg.Done()
	}()

	wg.Wait()

	return ctx.JSON(formJsonBody(
		BotListServicesResponse{
			Services:    responses,
			LastUpdated: time.Now().UnixMilli(),
		},
		true,
	))
}
