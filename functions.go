package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"time"
)

var conn *pgxpool.Pool

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
