package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"time"
)

// postGuildCountRoute is a function that used to post guild count information to all bot lists and be persisted in the database.
//
//	@Summary		Post guild stats to bot lists and persist them in the database.
//	@Description	The guild count and shard count are persisted to the database then posted to all active bot lists set in the config.
//	@tags			General
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	ResponseHTTP{data=GuildCountResponse}
//	@Failure		503				{object}	ResponseHTTPError{}
//
//	@Param			Authorization	header		string					true	"The required API key"
//
//	@Param			request			body		GuildCountRequestBody	true	"The request body to pass in."
//
//	@Router			/api/v1/guilds [post]
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

// getGuildCountRoute is a function that returns the most recently committed guild count in the database.
//
//	@Summary		Get the recent guild count from the database.
//	@Description	The most recently posted guild and shard count in the database is returned as well as the timestamp of when this data was committed. This data reflects the guild count on the active bot lists.
//	@tags			General
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	ResponseHTTP{data=GuildCountResponse}
//	@Failure		500				{object}	ResponseHTTPError{}
//
//	@Param			Authorization	header		string	true	"The required API key"
//
//	@Router			/api/v1/guilds [get]
func getGuildCountRoute(ctx *fiber.Ctx) error {
	var guildCount int64
	var shardCount int64
	var createdAt time.Time

	query := "select guild_count, shard_count, created_at from guildcount where shard_count is not null order by created_at desc"
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

// getBotListServicesRoute is a function to get an overview of all active lists the bot is on.
//
//	@Summary		Get all active lists the bot is on.
//	@Description	This function returns the timestamp of when guild stats were lasted committed to the database as well as an overview of all information from bot lists that are marked active via the config.
//	@tags			General
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	ResponseHTTP{data=BotListServicesResponse}
//	@Failure		500				{object}	ResponseHTTPError{}
//
//	@Param			Authorization	header		string	true	"The required API key"
//
//	@Router			/api/v1/services [get]
func getBotListServicesRoute(ctx *fiber.Ctx) error {
	responses, errors := fetchBotListServiceData()
	if len(errors) > 0 {
		return handleBotListErrors(ctx, errors)
	}

	var timestamp time.Time
	query := "select created_at from guildcount where shard_count is not null order by created_at desc"
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

// getSingleBotListServiceRoute is a function to get an overview of the specific bot list the bot is on.
//
//	@Summary		Get a single list the bot is on.
//	@Description	This function returns the timestamp of when guild stats were lasted committed to the database as well as an overview of the specific bot list the bot is on.
//	@tags			General
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	ResponseHTTP{data=BotListServicesResponse}
//
//	@Failure		400				{object}	ResponseHTTPError{data=InvalidServiceError}
//
//	@Failure		500				{object}	ResponseHTTPError{data=DefaultFiberError}
//
//	@Param			Authorization	header		string	true	"The required API key"
//
//	@Param			service			path		string	true	"The bot list service to get information from."
//
//	@Router			/api/v1/services/{service} [get]
func getSingleBotListServiceRoute(ctx *fiber.Ctx) error {
	service := ctx.Params("service")
	activeServices := getActiveServices()
	var services []BotListServiceResponse

	for _, s := range activeServices {
		if s == service {
			config := getServiceConfig(service)
			client := &http.Client{Timeout: time.Second * 30}
			data, err := fetchStats(client, config)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}

			services = append(services, *data)
			break
		}
	}

	if len(services) < 1 {
		msg := fmt.Sprintf("The service '%s' is not a valid service.", service)
		return fiber.NewError(fiber.StatusBadRequest, msg)
	}

	var timestamp time.Time
	query := "select created_at from guildcount order by created_at"
	queryRowError := queryRow(query, &timestamp)
	if queryRowError != nil {
		return fiber.NewError(fiber.StatusInternalServerError, queryRowError.Error())
	}

	return ctx.JSON(formJsonBody(
		BotListServicesResponse{
			Services:    services,
			LastUpdated: timestamp.UnixMilli(),
		},
		true,
	))
}
