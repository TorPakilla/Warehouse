package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	type ProductUnit struct {
		Type      string `json:"type"`
		ProductID string `gorm:"foreignKey:ProductID" json:"productid"`
	}

	var req ProductUnit
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Bad Request"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"type":      true,
		"productid": true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	CheckType := map[string]bool{
		"Pallet": true,
		"Box":    true,
		"Pieces": true,
	}

	if !CheckType[req.Type] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Type. Allowed Pallet, Box, Pieces"})
	}

	var ConverRate *int
	if req.Type == "Pallet" {
		Rate := 30
		ConverRate = &Rate
	} else if req.Type == "Box" {
		Rate := 12
		ConverRate = &Rate
	} else {
		ConverRate = nil
	}

	var product Models.Product
	if err := db.First(&product, "product_id = ?", req.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	productUnit := Models.ProductUnit{
		Type:        req.Type,
		ConversRate: ConverRate,
		ProductID:   req.ProductID,
	}

	if err := db.Create(&productUnit).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error"})
	}

	return c.Status(201).JSON(fiber.Map{"Data": productUnit})
}

func LookProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	var productUnit []Models.ProductUnit
	if err := db.Find(&productUnit).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Internal Server Error"})
	}
	return c.JSON(fiber.Map{"Product Units": productUnit})
}

func FindProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var productUnit Models.ProductUnit
	if err := db.First(&productUnit, "product_unit_id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product Unit not found"})
	}
	return c.JSON(fiber.Map{"Product Unit": productUnit})
}

func DeleteProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var productUnit Models.ProductUnit
	if err := db.Where("product_unit_id = ?", id).First(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product Unit not found"})
	}

	if err := db.Delete(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete Product Unit"})
	}

	return c.JSON(fiber.Map{"Delete": "Success", "Data": productUnit})
}

func UpdateProductUnit(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var productUnit Models.ProductUnit
	if err := db.Where("product_unit_id = ?", id).First(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product Unit not found"})
	}

	type ProductUnit struct {
		Type      string `json:"type"`
		ProductID string `json:"productid"`
	}

	var req ProductUnit
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"type":      true,
		"productid": true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	CheckType := map[string]bool{
		"Pallet": true,
		"Box":    true,
	}

	if !CheckType[req.Type] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Type. Allowed Pallet, Box, Pieces"})
	}

	var ConverRate *int
	if req.Type == "Pallet" {
		Rate := 12
		ConverRate = &Rate
	} else if req.Type == "Box" {
		Rate := 30
		ConverRate = &Rate
	} else {
		ConverRate = nil
	}

	productUnit.Type = req.Type
	productUnit.ConversRate = ConverRate
	productUnit.ProductID = req.ProductID

	if err := db.Save(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update Product Unit"})
	}

	return c.JSON(fiber.Map{"Update": "Success"})
}

func ProductUnitRouter(app *fiber.App, db *gorm.DB) {
	app.Use(func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "God" && role != "Manager" && role != "Stock" {
			return c.Next()
		}

		if role != "Account" && role != "Audit" {
			if c.Method() != "GET" {
				return c.Next()
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	})

	app.Post("/ProductUnit", func(c *fiber.Ctx) error {
		return AddProductUnit(db, c)
	})

	app.Get("/ProductUnit/:id", func(c *fiber.Ctx) error {
		return FindProductUnit(db, c)
	})

	app.Get("/ProductUnit", func(c *fiber.Ctx) error {
		return LookProductUnit(db, c)
	})

	app.Delete("/ProductUnit/:id", func(c *fiber.Ctx) error {
		return DeleteProductUnit(db, c)
	})

	app.Put("/ProductUnit/:id", func(c *fiber.Ctx) error {
		return UpdateProductUnit(db, c)
	})
}
