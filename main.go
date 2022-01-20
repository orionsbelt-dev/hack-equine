package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"hack/horses"
	"hack/riders"
	"hack/rides"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func setup() error {
	db, err := sql.Open("mysql", "tcp(localhost:3306)/?parseTime=true")
	if err != nil {
		return errors.New("Failed to connect to database: " + err.Error())
	}
	app := fiber.New()
	app.Use(logger.New())
	// TODO: add middleware for api key authentication

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world!")
	})

	app.Post("/horse", func(c *fiber.Ctx) error {
		var horse horses.Horse
		err := c.BodyParser(&horse)
		if err != nil {
			msg := "Failed to parse horse: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = horse.Save(db)
		if err != nil {
			msg := "Failed to save horse: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"horse": horse,
		})
	})

	app.Get("/horses", func(c *fiber.Ctx) error {
		horses, err := horses.GetHorses(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get horses: " + err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"horses": horses,
		})
	})

	app.Post("/rider", func(c *fiber.Ctx) error {
		var rider riders.Rider
		err := c.BodyParser(&rider)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse rider: " + err.Error(),
			})
		}
		err = rider.Save(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save rider: " + err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"rider": rider,
		})
	})

	app.Get("/riders", func(c *fiber.Ctx) error {
		riders, err := riders.GetRiders(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get riders: " + err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"riders": riders,
		})
	})

	app.Post("/ride", func(c *fiber.Ctx) error {
		var ride rides.Ride
		err := c.BodyParser(&ride)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse ride: " + err.Error(),
			})
		}
		err = ride.Save(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save ride: " + err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"ride": ride,
		})
	})

	return app.Listen(":8000")
}

func main() {
	err := setup()
	if err != nil {
		log.Fatal("Failed to setup app: " + err.Error())
	}
}
