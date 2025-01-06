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

	inventory := Models.Inventory{
		ProductUnitID: req.ProductUnitID,
		BrancheID:     req.BrancheID,
		Quantity:      req.Quantity,
		Price:         req.Price,
	}

	if err := db.Create(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inventory: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Inventory created successfully", "data": inventory})
}

func LookInventory(db *gorm.DB, c *fiber.Ctx) error {
	var inventories []Models.Inventory
	if err := db.Find(&inventories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลสินค้าทั้งหมดได้: " + err.Error()})
	}

	type InventoryBox struct {
		InventoryID string  `json:"inventoryid"`
		Type        string  `json:"type"`
		ConversRate *int    `json:"conversrate"`
		Quantity    int     `json:"quantity"`
		Box         int     `json:"Box"`
		Price       float64 `json:"price"`
		TotalPrice  float64 `json:"totalprice"`
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
		inventoryBox := InventoryBox{
			InventoryID: inventory.InventoryID,
			Type:        productUnit.Type,
			ConversRate: productUnit.ConversRate,
			Quantity:    inventory.Quantity,
			Box:         Box,
			Price:       inventory.Price,
			TotalPrice:  totalPrice,
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
		Type        string  `json:"type"`
		ConversRate *int    `json:"conversrate"`
		Quantity    int     `json:"quantity"`
		Box         int     `json:"box"`
		Price       float64 `json:"price"`
		TotalPrice  float64 `json:"totalprice"`
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

	inventoryBox := InventoryBox{
		Type:        productUnit.Type,
		ConversRate: productUnit.ConversRate,
		Quantity:    inventory.Quantity,
		Box:         Box,
		Price:       inventory.Price,
		TotalPrice:  totalPrice,
	}

	return c.JSON(fiber.Map{"data": inventoryBox})
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

	inventory.ProductUnitID = req.ProductUnitID
	inventory.BrancheID = req.BrancheID
	inventory.Quantity = req.Quantity
	inventory.Price = req.Price

	if err := db.Save(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Inventory updated successfully", "data": inventory})
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
