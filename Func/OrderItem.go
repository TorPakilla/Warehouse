package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	type OrderItemRequest struct {
		OrderID       string  `json:"orderid"`
		ProductUnitID string  `json:"productunitid"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unitprice"`
	}

	var req OrderItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	orderItem := Models.OrderItem{
		OrderID:       req.OrderID,
		ProductUnitID: req.ProductUnitID,
		Quantity:      req.Quantity,
		UnitPrice:     req.UnitPrice,
	}

	if err := db.Create(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New Order Item": orderItem})
}

func LookOrderItems(db *gorm.DB, c *fiber.Ctx) error {
	var orderItems []Models.OrderItem
	if err := db.Find(&orderItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find order items: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "Order Items", "Data": orderItems})
}

func FindOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var orderItem Models.OrderItem
	if err := db.Where("order_item_id = ?", id).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order Item not found"})
	}
	return c.JSON(fiber.Map{"This": "Order Item", "Data": orderItem})
}

func DeleteOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var orderItem Models.OrderItem
	if err := db.Where("order_item_id = ?", id).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order Item not found"})
	}
	if err := db.Delete(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete order item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": "Succeed"})
}

func UpdateOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var orderItem Models.OrderItem
	if err := db.Where("order_item_id = ?", id).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order Item not found"})
	}

	type OrderItemRequest struct {
		OrderID       string  `json:"orderid"`
		ProductUnitID string  `json:"productunitid"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unitprice"`
	}

	var req OrderItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	orderItem.OrderID = req.OrderID
	orderItem.ProductUnitID = req.ProductUnitID
	orderItem.Quantity = req.Quantity
	orderItem.UnitPrice = req.UnitPrice

	if err := db.Save(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
}

func OrderItemRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/OrderItems", func(c *fiber.Ctx) error {
		return LookOrderItems(db, c)
	})
	app.Get("/OrderItems/:id", func(c *fiber.Ctx) error {
		return FindOrderItem(db, c)
	})
	app.Post("/OrderItems", func(c *fiber.Ctx) error {
		return AddOrderItem(db, c)
	})
	app.Put("/OrderItems/:id", func(c *fiber.Ctx) error {
		return UpdateOrderItem(db, c)
	})
	app.Delete("/OrderItems/:id", func(c *fiber.Ctx) error {
		return DeleteOrderItem(db, c)
	})
}
