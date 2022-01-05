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
	var responses []BotListServiceResponse

	var wg sync.WaitGroup
	wg.Add(5)

	wgErrors := make(chan error)
	wgDone := make(chan bool)

	client := &http.Client{Timeout: time.Second * 30}

	go fetchStats(client, "topgg", &responses, &wg, &wgErrors)
	go fetchStats(client, "botsgg", &responses, &wg, &wgErrors)
	go fetchStats(client, "dlspace", &responses, &wg, &wgErrors)
	go fetchStats(client, "dbl", &responses, &wg, &wgErrors)
	go fetchStats(client, "discords", &responses, &wg, &wgErrors)

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-wgErrors:
		close(wgErrors)
		panic(err)
	}

	return ctx.JSON(formJsonBody(
		BotListServicesResponse{
			Services:    responses,
			LastUpdated: time.Now().UnixMilli(),
		},
		true,
	))
}
