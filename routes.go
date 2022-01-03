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
	var responses []*BotListServiceResponse

	var wg sync.WaitGroup
	wg.Add(5)

	client := &http.Client{Timeout: time.Second * 30}

	// top.gg
	go func() {
		body, reqErr := getDataFromBotListService(client, "topgg")
		if reqErr != nil {
			fiber.NewError(fiber.StatusInternalServerError, reqErr.Error())
		}

		responses = append(responses, &BotListServiceResponse{
			Id:         services.Get("services.topgg.id").(int64),
			ShortName:  services.Get("services.topgg.short_name").(string),
			Url:        services.Get("services.topgg.get_stats_url").(string),
			GuildCount: int64(body["server_count"].(float64)),
		})

		wg.Done()
	}()

	// discord.bots.gg
	go func() {
		body, reqErr := getDataFromBotListService(client, "botsgg")
		if reqErr != nil {
			fiber.NewError(fiber.StatusInternalServerError, reqErr.Error())
		}

		responses = append(responses, &BotListServiceResponse{
			Id:         services.Get("services.botsgg.id").(int64),
			ShortName:  services.Get("services.botsgg.short_name").(string),
			Url:        services.Get("services.botsgg.get_stats_url").(string),
			GuildCount: int64(body["guildCount"].(float64)),
		})

		wg.Done()
	}()

	// discordlist.space
	go func() {
		body, reqErr := getDataFromBotListService(client, "dlspace")
		if reqErr != nil {
			fiber.NewError(fiber.StatusInternalServerError, reqErr.Error())
		}

		responses = append(responses, &BotListServiceResponse{
			Id:         services.Get("services.dlspace.id").(int64),
			ShortName:  services.Get("services.dlspace.short_name").(string),
			Url:        services.Get("services.dlspace.get_stats_url").(string),
			GuildCount: int64(body["serverCount"].(float64)),
		})

		wg.Done()
	}()

	// discordbotlist.com
	go func() {
		body, reqErr := getDataFromBotListService(client, "dbl")
		if reqErr != nil {
			fiber.NewError(fiber.StatusInternalServerError, reqErr.Error())
		}

		fmt.Printf("%s", body["stats"].(map[string]interface{})["guilds"])

		responses = append(responses, &BotListServiceResponse{
			Id:         services.Get("services.dbl.id").(int64),
			ShortName:  services.Get("services.dbl.short_name").(string),
			Url:        services.Get("services.dbl.get_stats_url").(string),
			GuildCount: int64(body["stats"].(map[string]interface{})["guilds"].(float64)),
		})

		wg.Done()
	}()

	// discords.com
	go func() {
		body, reqErr := getDataFromBotListService(client, "discords")
		if reqErr != nil {
			fiber.NewError(fiber.StatusInternalServerError, reqErr.Error())
		}

		responses = append(responses, &BotListServiceResponse{
			Id:         services.Get("services.discords.id").(int64),
			ShortName:  services.Get("services.discords.short_name").(string),
			Url:        services.Get("services.discords.get_stats_url").(string),
			GuildCount: int64(body["server_count"].(float64)),
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
