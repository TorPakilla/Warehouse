package Database

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddBranche(db *gorm.DB, c *fiber.Ctx) error {
	var branche Models.Branches
	if err := c.BodyParser(&branche); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format: " + err.Error(),
		})
	}

	if err := db.Create(&branche).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create branche: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(branche)
}
