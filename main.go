package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"hack/barns"
	"hack/horses"
	"hack/riders"
	"hack/rides"
	"hack/users"
	"hack/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v4/stytch"
	"github.com/stytchauth/stytch-go/v4/stytch/stytchapi"
)

func setup() error {
	godotenv.Load()
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return errors.New("API_KEY not set")
	}
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		return errors.New("Failed to connect to database: " + err.Error())
	}

	// stytch
	client, err := stytchapi.NewAPIClient(
		stytch.EnvTest,
		os.Getenv("STYTCH_PROJECT_ID"),
		os.Getenv("STYTCH_PROJECT_SECRET"),
	)
	if err != nil {
		return errors.New("Failed to create stytch client: " + err.Error())
	}

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(func(c *fiber.Ctx) error {
		providedKey := c.Get("x-api-key")
		if providedKey != apiKey {
			msg := "invalid api key"
			fmt.Println(msg)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world!")
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		var user users.User
		err := c.BodyParser(&user)
		if err != nil {
			msg := "error parsing user: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = user.Login(client, db)
		if err != nil {
			msg := "error logging in user: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(user)
	})

	app.Post("/signup", func(c *fiber.Ctx) error {
		var user users.User
		err := c.BodyParser(&user)
		if err != nil {
			msg := "error parsing user: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = user.Signup(client, db)
		if err != nil {
			msg := "error signing up user: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(user)
	})

	app.Post("/authenticate", func(c *fiber.Ctx) error {
		var userAuth users.UserAuth
		err := c.BodyParser(&userAuth)
		if err != nil {
			msg := "error parsing user auth: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		user, err := userAuth.AuthenticatePasscode(client, db)
		if err != nil {
			msg := "error authenticating passcode: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(*user)
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
		userID := c.Params("userID")
		barns, err := barns.GetBarnsByUserID(userID, db)
		if err != nil {
			msg := "Failed to get barns: " + err.Error()
			fmt.Println(msg)
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
			msg := "Failed to parse ride: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = ride.Save(db)
		if err != nil {
			msg := "Failed to save ride: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"ride": ride,
		})
	})

	app.Put("/ride/cancel", func(c *fiber.Ctx) error {
		var ride rides.Ride
		err := c.BodyParser(&ride)
		if err != nil {
			msg := "Failed to parse ride: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = ride.Cancel(db)
		if err != nil {
			msg := "Failed to cancel ride: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"ride": ride,
		})
	})

	app.Post("/event/type", func(c *fiber.Ctx) error {
		var eventType rides.EventType
		err := c.BodyParser(&eventType)
		if err != nil {
			msg := "Failed to parse event type: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = eventType.Save(db)
		if err != nil {
			msg := "Failed to save event type: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"event_type": eventType,
		})
	})

	app.Get("/event/types", func(c *fiber.Ctx) error {
		types, err := rides.ListEventTypes(db)
		if err != nil {
			msg := "Failed to get event types: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"event_types": types,
		})
	})

	app.Post("/schedule", func(c *fiber.Ctx) error {
		var schedule rides.Schedule
		err := c.BodyParser(&schedule)
		if err != nil {
			msg := "Failed to parse schedule: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = schedule.Save(db)
		if err != nil {
			msg := "Failed to save schedule: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"success": true,
		})
	})

	app.Delete("/schedule/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			msg := "Failed to parse schedule id: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = rides.DeleteSchedule(id, db)
		if err != nil {
			msg := "Failed to delete schedule: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		c.Status(fiber.StatusOK)
		return nil
	})

	app.Get("/barn/:barnID/rides/:date", func(c *fiber.Ctx) error {
		date, err := time.Parse("2006-01-02", c.Params("date"))
		if err != nil {
			msg := "Failed to parse date: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		barnID, err := strconv.ParseInt(c.Params("barnID"), 10, 64)
		if err != nil {
			msg := "Failed to parse barn ID: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		rides, err := rides.GetScheduleByDay(barnID, utils.Date{Time: date}, db)
		if err != nil {
			msg := "Failed to get ride schedule: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"rides": rides,
		})
	})

	app.Get("/horse/:id/schedule/:date", func(c *fiber.Ctx) error {
		date, err := time.Parse("2006-01-02", c.Params("date"))
		if err != nil {
			msg := "Failed to parse date: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			msg := "Failed to parse horse ID: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		schedule, err := rides.GetHorseScheduleByDay(id, utils.Date{Time: date}, db)
		if err != nil {
			msg := "Failed to get schedule: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"schedule": schedule,
		})
	})

	app.Get("/barn/:barnID/recurring", func(c *fiber.Ctx) error {
		barnID, err := strconv.ParseInt(c.Params("barnID"), 10, 64)
		if err != nil {
			msg := "Failed to parse barn ID: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		schedules, err := rides.ListSchedules(barnID, db)
		if err != nil {
			msg := "Failed to list recurring schedules: " + err.Error()
			fmt.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(fiber.Map{
			"schedules": schedules,
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
