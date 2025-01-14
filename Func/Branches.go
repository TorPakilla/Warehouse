package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddBranches(db *gorm.DB, c *fiber.Ctx) error {
	type Branches struct {
		BName    string `json:"bname" validate:"required"`
		Location string `json:"location" validate:"required"`
	}

	var req Branches
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	if req.BName == "" || req.Location == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BName and Location are required"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"bname":    true,
		"location": true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
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

func LookBranch(db *gorm.DB, c *fiber.Ctx) error {
	var branches []Models.Branches
	if err := db.Find(&branches).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error"})
	}
	return c.JSON(fiber.Map{"Branches": branches}) // ส่งข้อมูลใน key "Branches"
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

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"bname":    true,
		"location": true,
	}

	for key, value := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Worng: " + key})
		}

		switch key {
		case "bname":
			if v, ok := value.(string); ok {
				branche.BName = v
			}
		case "location":
			if v, ok := value.(string); ok {
				branche.Location = v
			}
		}
	}

	if err := db.Save(&branche).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cant Update: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Update Succeed"})
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
		return LookBranch(db, c)
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
