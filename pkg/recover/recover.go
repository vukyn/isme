package recover

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/vukyn/kuery/log"
)

func NewFiberRecover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Catch panics
		defer func() {
			if err := recover(); err != nil {
				switch err := err.(type) {
				case error:
					log.New().Errorf("Panic recovered: %v", err)
				case string:
					log.New().Errorf("Panic recovered: %v", errors.New(err))
				}
			}
		}()
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong, please try again later!",
		})
		return c.Next()
	}
}
