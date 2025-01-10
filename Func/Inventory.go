package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddInventory(db *gorm.DB, c *fiber.Ctx) error {
	type InventoryRequest struct {
		ProductUnitID string  `json:"productunitid"`
		BrancheID     string  `json:"brancheid"`
		Quantity      int     `json:"quantity"`
		Price         float64 `json:"price"`
	}

	var req InventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	if req.Quantity < 0 || req.Price < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quantity and Price must be greater"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"productunitid": true,
		"brancheid":     true,
		"quantity":      true,
		"price":         true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	inventory := Models.Inventory{
		ProductUnitID: req.ProductUnitID,
		BrancheID:     req.BrancheID,
		Quantity:      req.Quantity,
		Price:         req.Price,
	}

	if err := db.Create(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inventory: " + err.Error()})
	}

	var quantityStatus string
	if inventory.Quantity < 1000 {
		quantityStatus = "Low"
	} else if inventory.Quantity < 5000 {
		quantityStatus = "Medium"
	} else {
		quantityStatus = "High"
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Inventory created successfully", "data": inventory, "quantity_status": quantityStatus})
}

func UpdateInventory(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var inventory Models.Inventory
	if err := db.Where("inventory_id = ?", id).First(&inventory).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory not found"})
	}

	type InventoryRequest struct {
		ProductUnitID string  `json:"productunitid"`
		Quantity      int     `json:"quantity"`
		Price         float64 `json:"price"`
		BrancheID     string  `json:"brancheid"`
	}

	var req InventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	if req.Quantity < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quantity must be greater"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"productunitid": true,
		"brancheid":     true,
		"quantity":      true,
		"price":         true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	if req.Quantity > inventory.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Need more quantity"})
	}

	inventory.ProductUnitID = req.ProductUnitID
	inventory.BrancheID = req.BrancheID
	inventory.Quantity = req.Quantity
	inventory.Price = req.Price

	if err := db.Save(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory: " + err.Error()})
	}

	var quantityStatus string
	if inventory.Quantity < 1000 {
		quantityStatus = "Low"
	} else if inventory.Quantity < 5000 {
		quantityStatus = "Medium"
	} else {
		quantityStatus = "High"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Inventory updated successfully", "data": inventory, "quantity_status": quantityStatus})
}

func LookInventory(db *gorm.DB, c *fiber.Ctx) error {
	var inventories []Models.Inventory
	if err := db.Find(&inventories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cant LookInventory Data: " + err.Error()})
	}

	type InventoryBox struct {
		InventoryID    string  `json:"inventoryid"`
		Type           string  `json:"type"`
		ConversRate    *int    `json:"conversrate"`
		Quantity       int     `json:"quantity"`
		Box            int     `json:"Box"`
		Price          float64 `json:"price"`
		TotalPrice     float64 `json:"totalprice"`
		QuantityStatus string  `json:"quantitystatus"`
	}

	var result []InventoryBox
	for _, inventory := range inventories {
		var productUnit Models.ProductUnit
		if err := db.Where("product_unit_id = ?", inventory.ProductUnitID).First(&productUnit).Error; err != nil {
			continue
		}

		var Box int
		if productUnit.ConversRate != nil {
			Box = inventory.Quantity * *productUnit.ConversRate
		}

		totalPrice := float64(inventory.Quantity) * inventory.Price
		var quantityStatus string
		if inventory.Quantity < 1000 {
			quantityStatus = "Low"
		} else if inventory.Quantity < 5000 {
			quantityStatus = "Medium"
		} else {
			quantityStatus = "High"
		}

		inventoryBox := InventoryBox{
			InventoryID:    inventory.InventoryID,
			Type:           productUnit.Type,
			ConversRate:    productUnit.ConversRate,
			Quantity:       inventory.Quantity,
			Box:            Box,
			Price:          inventory.Price,
			TotalPrice:     totalPrice,
			QuantityStatus: quantityStatus,
		}
		result = append(result, inventoryBox)
	}

	return c.JSON(fiber.Map{"data": result})
}

func FindInventory(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var inventory Models.Inventory
	if err := db.Where("inventory_id = ?", id).First(&inventory).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory not found"})
	}

	type InventoryBox struct {
		Type           string  `json:"type"`
		ConversRate    *int    `json:"conversrate"`
		Quantity       int     `json:"quantity"`
		Box            int     `json:"box"`
		Price          float64 `json:"price"`
		TotalPrice     float64 `json:"totalprice"`
		QuantityStatus string  `json:"quantitystatus"`
	}

	var productUnit Models.ProductUnit
	if err := db.Where("product_unit_id = ?", inventory.ProductUnitID).First(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ProductUnit not found"})
	}

	var Box int
	if productUnit.ConversRate != nil {
		Box = inventory.Quantity * *productUnit.ConversRate
	}

	totalPrice := float64(inventory.Quantity) * inventory.Price
	var quantityStatus string
	if inventory.Quantity < 1000 {
		quantityStatus = "Low"
	} else if inventory.Quantity < 5000 {
		quantityStatus = "Medium"
	} else {
		quantityStatus = "High"
	}

	inventoryBox := InventoryBox{
		Type:           productUnit.Type,
		ConversRate:    productUnit.ConversRate,
		Quantity:       inventory.Quantity,
		Box:            Box,
		Price:          inventory.Price,
		TotalPrice:     totalPrice,
		QuantityStatus: quantityStatus,
	}

	return c.JSON(fiber.Map{"data": inventoryBox})
}

func DeleteInventory(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var inventory Models.Inventory
	if err := db.Where("inventory_id = ?", id).First(&inventory).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory not found"})
	}
	if err := db.Delete(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete inventory: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Inventory deleted successfully"})
}

func InventoryRoutes(app *fiber.App, db *gorm.DB) {
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

	app.Get("/Inventory", func(c *fiber.Ctx) error {
		return LookInventory(db, c)
	})

	app.Get("/Inventory/:id", func(c *fiber.Ctx) error {
		return FindInventory(db, c)
	})

	app.Post("/Inventory", func(c *fiber.Ctx) error {
		return AddInventory(db, c)
	})

	app.Put("/Inventory/:id", func(c *fiber.Ctx) error {
		return UpdateInventory(db, c)
	})

	app.Delete("/Inventory/:id", func(c *fiber.Ctx) error {
		return DeleteInventory(db, c)
	})
}
