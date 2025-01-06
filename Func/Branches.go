package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddBranches(db *gorm.DB, c *fiber.Ctx) error {
	type Branches struct {
		BName    string `json:"bname"`
		Location string `json:"location"`
	}

	var req Branches
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	branche := Models.Branches{
		BName:    req.BName,
		Location: req.Location,
	}

	if err := db.Create(&branche).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create branche: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New": branche})
}

func LookBranches(db *gorm.DB, c *fiber.Ctx) error {
	var branches []Models.Branches
	if err := db.Find(&branches).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find branches: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "Branches", "Data": branches})
}

func FindBranches(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var branche Models.Branches
	if err := db.Where("branche_id = ?", id).First(&branche).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branche not found"})
	}

	return c.JSON(fiber.Map{"This": "Branche", "Data": branche})
}

func DeleteBranches(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var branche Models.Branches
	if err := db.Where("branche_id = ?", id).First(&branche).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branche not found"})
	}

	if err := db.Delete(&branche).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete branche: " + err.Error()})
	}

	return c.JSON(fiber.Map{"Delete": "Succeed"})
}

func UpdateBranches(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var branche Models.Branches
	if err := db.Where("branche_id = ?", id).First(&branche).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branche not found"})
	}

	type Branches struct {
		BName    string `json:"bname"`
		Location string `json:"location"`
	}

	var req Branches
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	branche.BName = req.BName
	branche.Location = req.Location

	if err := db.Save(&branche).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update branche: " + err.Error()})
	}

	return c.JSON(fiber.Map{"Update": "Succeed"})
}

func BranchesRoutes(app *fiber.App, db *gorm.DB) {
	app.Use(func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "God" && role != "Manager" {
			return c.Next()
		}

		if role != "Stock" && role != "Account" && role != "Audit" {
			if c.Method() != "GET" && c.Method() != "UPDATE" {
				return c.Next()
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	})

	app.Post("/Branches", func(c *fiber.Ctx) error {
		return AddBranches(db, c)
	})

	app.Get("/Branches", func(c *fiber.Ctx) error {
		return LookBranches(db, c)
	})

	app.Get("/Branches/:id", func(c *fiber.Ctx) error {
		return FindBranches(db, c)
	})

	app.Delete("/Branches/:id", func(c *fiber.Ctx) error {
		return DeleteBranches(db, c)
	})

	app.Put("/Branches/:id", func(c *fiber.Ctx) error {
		return UpdateBranches(db, c)
	})
}
