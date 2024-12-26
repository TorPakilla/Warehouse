package Func

import (
	"Api/Models"

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

	product := Models.Product{
		ProductName: req.ProductName,
		Description: req.Description,
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
	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	return c.JSON(fiber.Map{"This": "Product", "Data": product})
}

func DeleteProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
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

	var product Models.Product
	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
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
	app.Post("/Product", func(c *fiber.Ctx) error {
		return AddProduct(db, c)
	})
	app.Get("/Product", func(c *fiber.Ctx) error {
		return LookProduct(db, c)
	})
	app.Get("/Product", func(c *fiber.Ctx) error {
		return FindProduct(db, c)
	})
	app.Delete("/Product", func(c *fiber.Ctx) error {
		return DeleteProduct(db, c)
	})
	app.Put("/Product", func(c *fiber.Ctx) error {
		return UpdateProduct(db, c)
	})
}
