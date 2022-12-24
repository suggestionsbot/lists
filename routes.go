package main

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func getRootRoute(ctx *fiber.Ctx) error {
	return ctx.JSON(formJsonBody(
		fiber.Map{"message": "Hello world!"},
		true,
	))
}

func postGuildCountRoute(ctx *fiber.Ctx) error {
	guild := new(GuildCountRequestBody)

	if err := ctx.BodyParser(guild); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	errors := validateGuildCount(*guild)
	if errors != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(formJsonBody(errors, false))
	}

	if !guild.DryRun {
		query := "insert into guildcount(guild_count, shard_count) values ($1, $2) returning *"
		_, execErr := execQuery(query, guild.Guilds, guild.Shards)
		if execErr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, execErr.Error())
		}

		postErrors := postStatsToBotLists(guild.Guilds, guild.Shards)
		if len(postErrors) > 0 {
			return handleBotListErrors(ctx, postErrors)
		}
	}

	return ctx.JSON(formJsonBody(GuildCountResponse{
		Guilds:    guild.Guilds,
		Shards:    guild.Shards,
		DryRun:    guild.DryRun,
		Timestamp: time.Now().UnixMilli(),
	}, true))
}

func getGuildCountRoute(ctx *fiber.Ctx) error {
	var guildCount int64
	var shardCount int64
	var createdAt time.Time

	query := "select guild_count, shard_count, created_at from guildcount order by created_at desc"
	err := queryRow(query, &guildCount, &shardCount, &createdAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(formJsonBody(
		GuildCountResponse{
			Guilds:    guildCount,
			Shards:    shardCount,
			Timestamp: createdAt.UnixMilli(),
		}, true,
	))
}

func getBotListServicesRoute(ctx *fiber.Ctx) error {
	responses, errors := fetchBotListServiceData()
	if len(errors) > 0 {
		return handleBotListErrors(ctx, errors)
	}

	var timestamp time.Time
	query := "select created_at from guildcount order by created_at desc"
	queryRowError := queryRow(query, &timestamp)
	if queryRowError != nil {
		return fiber.NewError(fiber.StatusInternalServerError, queryRowError.Error())
	}

	return ctx.JSON(formJsonBody(
		BotListServicesResponse{
			Services:    responses,
			LastUpdated: timestamp.UnixMilli(),
		},
		true,
	))
}
