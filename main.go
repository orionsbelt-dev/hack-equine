package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"hack/barns"
	"hack/horses"
	"hack/riders"
	"hack/rides"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setup() error {
	godotenv.Load()
	db, err := sql.Open("mysql", "tcp(localhost:3306)/?parseTime=true")
	if err != nil {
		return errors.New("Failed to connect to database: " + err.Error())
	}
	app := fiber.New()
	app.Use(logger.New())
	app.Use(func(c *fiber.Ctx) error {
		providedKey := c.Get("x-api-key")
		expectedKey := os.Getenv("API_KEY")
		if providedKey != expectedKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid api key",
			})
		}
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world!")
	})

	app.Post("/barn", func(c *fiber.Ctx) error {
		logger := c.Context().Logger()
		type barnRequest struct {
			Name   string `json:"name"`
			UserID string `json:"user_id"`
		}
		var req barnRequest
		err := c.BodyParser(&req)
		if err != nil {
			msg := "Failed to parse request body: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		barn := barns.Barn{
			Name: req.Name,
		}
		err = barn.Save(req.UserID, db)
		if err != nil {
			msg := "Failed to save barn: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"barn": barn,
		})
	})

	app.Get("/user/:userID/barns", func(c *fiber.Ctx) error {
		logger := c.Context().Logger()
		userID := c.Params("userID")
		barns, err := barns.GetBarnsByUserID(userID, db)
		if err != nil {
			msg := "Failed to get barns: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"barns": barns,
		})
	})

	app.Get("/barn/:barnID/horses", func(c *fiber.Ctx) error {
		logger := c.Context().Logger()
		barnID, err := strconv.ParseInt(c.Params("barnID"), 10, 64)
		if err != nil {
			msg := "Failed to parse barn ID: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		horses, err := horses.GetHorsesByBarnID(barnID, db)
		if err != nil {
			msg := "Failed to get horses: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"horses": horses,
		})
	})

	app.Get("/barn/:barnID/riders", func(c *fiber.Ctx) error {
		logger := c.Context().Logger()
		barnID, err := strconv.ParseInt(c.Params("barnID"), 10, 64)
		if err != nil {
			msg := "Failed to parse barn ID: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		riders, err := riders.GetRidersByBarnID(barnID, db)
		if err != nil {
			msg := "Failed to get riders: " + err.Error()
			logger.Printf(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"riders": riders,
		})
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

	app.Post("/schedule", func(c *fiber.Ctx) error {
		var schedule rides.Schedule
		err := c.BodyParser(&schedule)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse schedule: " + err.Error(),
			})
		}
		err = schedule.Save(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save schedule: " + err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"success": true,
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
