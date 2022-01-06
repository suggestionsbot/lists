package main

import (
	"encoding/json"
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

	query := "insert into guildcount(guild_count, timestamp) values (%d, %d) returning *"
	_, execErr := execQuery(query, jsonObj.Count, jsonObj.Timestamp)
	if execErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, execErr.Error())
	}

	return ctx.JSON(formJsonBody(GuildCountResponse{
		Count: jsonObj.Count,
	}, true))
}

func getGuildCount(ctx *fiber.Ctx) error {
	var guildCount int64
	var timestamp int64

	query := "select guild_count, timestamp from guildcount order by timestamp desc"
	err := queryRow(query, &guildCount, &timestamp)
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
	responses, errors := fetchBotListServiceData()
	if len(errors) >= 1 {
		var data []interface{}
		for _, err := range errors {
			data = append(data, err)
		}

		ctx.Status(fiber.StatusInternalServerError)
		return ctx.JSON(formJsonBody(data, false))
	}

	var timestamp int64
	query := "select timestamp from guildcount order by timestamp desc"
	queryRowError := queryRow(query, &timestamp)
	if queryRowError != nil {
		return fiber.NewError(fiber.StatusInternalServerError, queryRowError.Error())
	}

	return ctx.JSON(formJsonBody(
		BotListServicesResponse{
			Services:    responses,
			LastUpdated: timestamp,
		},
		true,
	))
}
