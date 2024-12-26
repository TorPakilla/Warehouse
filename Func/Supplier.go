package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddSupplier(db *gorm.DB, c *fiber.Ctx) error {
	type SupplierRequest struct {
		Name        string  `json:"name"`
		PricePallet float64 `json:"pricepallet"`
		ProductID   string  `gorm:"foreignKey:ProductID" json:"productid"`
	}

	var req SupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	supplier := Models.Supplier{
		Name:        req.Name,
		PricePallet: req.PricePallet,
		ProductID:   req.ProductID,
	}

	if err := db.Create(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create supplier: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Supplier created successfully", "data": supplier})
}

func LookSuppliers(db *gorm.DB, c *fiber.Ctx) error {
	var suppliers []Models.Supplier
	if err := db.Find(&suppliers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch suppliers: " + err.Error()})
	}
	return c.JSON(fiber.Map{"data": suppliers})
}

func FindSupplier(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", id).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}
	return c.JSON(fiber.Map{"data": supplier})
}

func UpdateSupplier(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", id).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}

	type SupplierRequest struct {
		Name        string  `json:"name"`
		PricePallet float64 `json:"pricepallet"`
		ProductID   string  `gorm:"foreignKey:ProductID" json:"productid"`
	}

	var req SupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	supplier.Name = req.Name
	supplier.PricePallet = req.PricePallet
	supplier.ProductID = req.ProductID

	if err := db.Save(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update supplier: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier updated successfully", "data": supplier})
}

func DeleteSupplier(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", id).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}
	if err := db.Delete(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete supplier: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier deleted successfully"})
}

func SupplierRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/Supplier", func(c *fiber.Ctx) error {
		return LookSuppliers(db, c)
	})
	app.Get("/Supplier/:id", func(c *fiber.Ctx) error {
		return FindSupplier(db, c)
	})
	app.Post("/Supplier", func(c *fiber.Ctx) error {
		return AddSupplier(db, c)
	})
	app.Put("/Supplier/:id", func(c *fiber.Ctx) error {
		return UpdateSupplier(db, c)
	})
	app.Delete("/Supplier/:id", func(c *fiber.Ctx) error {
		return DeleteSupplier(db, c)
	})
}
