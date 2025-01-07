package Func

import (
	"Api/Models"

	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddProduct(db *gorm.DB, c *fiber.Ctx) error {
	type ProductRequest struct {
		ProductName string `json:"productname"`
		Description string `json:"description"`
	}

	var req ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"productname": true,
		"description": true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	product := Models.Product{
		ProductName: req.ProductName,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New": product})
}

func LookProduct(db *gorm.DB, c *fiber.Ctx) error {
	var products []Models.Product
	if err := db.Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find products: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "Product", "Data": products})
}

func FindProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	return c.JSON(fiber.Map{"This": "Product", "Data": product})
}

func DeleteProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	if err := db.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete product: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": product})
}

func UpdateProduct(db *gorm.DB, c *fiber.Ctx) error {
	type ProductRequest struct {
		ProductName string `json:"productname"`
		Description string `json:"description"`
	}

	id := c.Params("id")
	var req ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"productname": true,
		"description": true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	product.ProductName = req.ProductName
	product.Description = req.Description

	if err := db.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": product})
}

func ProductRouter(app fiber.Router, db *gorm.DB) {
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
	app.Post("/Product", func(c *fiber.Ctx) error {
		return AddProduct(db, c)
	})

	app.Get("/Product", func(c *fiber.Ctx) error {
		return LookProduct(db, c)
	})

	app.Get("/Product/:id", func(c *fiber.Ctx) error {
		return FindProduct(db, c)
	})

	app.Delete("/Product/:id", func(c *fiber.Ctx) error {
		return DeleteProduct(db, c)
	})

	app.Put("/Product/:id", func(c *fiber.Ctx) error {
		return UpdateProduct(db, c)
	})
}
