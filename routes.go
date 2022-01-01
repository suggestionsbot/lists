package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
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
