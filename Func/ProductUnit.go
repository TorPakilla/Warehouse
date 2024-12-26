package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	type ProductUnit struct {
		Type        string `json:"type"`
		ConversRate int    `json:"convers_rate"`
	}

	var req ProductUnit
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Bad Request"})
	}

	productUnit := Models.ProductUnit{
		Type:        req.Type,
		ConversRate: req.ConversRate,
	}

	db.Create(&productUnit)
	if err := db.Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error"})
	}

	return c.Status(201).JSON(fiber.Map{"Add": "Product", "Data": productUnit})

}

func LookProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	var productUnit []Models.ProductUnit
	if err := db.Find(&productUnit).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error"})
	}
	return c.JSON(fiber.Map{"This Product": productUnit})
}

func FindProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var productUnit Models.ProductUnit
	if err := db.First(&productUnit, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error"})
	}
	return c.JSON(fiber.Map{"This Product": productUnit})
}

func DeleteProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var productUnit Models.ProductUnit
	if err := db.Where("id = ?", id).First(productUnit).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branche not found"})
	}

	if err := db.Delete(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete branche: " + err.Error()})
	}

	return c.JSON(fiber.Map{"Delete": "Product", "Data": productUnit})
}

func UpdateProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var productUnit Models.ProductUnit
	if err := db.Where("id = ?", id).First(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branche not found"})
	}

	type ProductUnit struct {
		Type        string `json:"type"`
		ConversRate int    `json:"convers_rate"`
	}

	var req ProductUnit
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	productUnit.Type = req.Type
	productUnit.ConversRate = req.ConversRate

	if err := db.Save(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update : " + err.Error()})
	}

	return c.JSON(fiber.Map{"Update": "Succeed"})
}

func ProductUnit(app *fiber.App, db *gorm.DB) {
	app.Post("/ProductUnit", func(c *fiber.Ctx) error {
		return AddProductUnit(db, c)
	})

	app.Get("/ProductUnit", func(c *fiber.Ctx) error {
		return LookProductUnit(db, c)
	})

	app.Get("/ProductUnit", func(c *fiber.Ctx) error {
		return FindProductUnit(db, c)
	})

	app.Delete("/ProductUnit", func(c *fiber.Ctx) error {
		return DeleteProductUnit(db, c)
	})

	app.Put("/ProductUnit", func(c *fiber.Ctx) error {
		return UpdateProductUnit(db, c)
	})
}
