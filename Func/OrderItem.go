package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	type OrderItemRequest struct {
		OrderID     string  `json:"orderid" validate:"required"`
		ProductID   string  `json:"productid" validate:"required"`
		Quantity    int     `json:"quantity" validate:"required,min=1"`
		ConversRate float64 `json:"conversrate" validate:"required,min=0"`
	}

	var req OrderItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบ JSON ไม่ถูกต้อง: " + err.Error()})
	}

	orderItem := Models.OrderItem{
		OrderID:     req.OrderID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		ConversRate: req.ConversRate,
	}

	// สร้าง OrderItem
	if err := db.Create(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถสร้างรายการคำสั่งซื้อได้: " + err.Error()})
	}

	// อัปเดตยอดรวมของ Order
	if err := UpdateTotalAmount(db, req.OrderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถอัปเดตยอดรวมได้: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "รายการคำสั่งซื้อถูกสร้างสำเร็จ", "data": orderItem})
}

// ลบรายการคำสั่งซื้อ
func DeleteOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var orderItem Models.OrderItem
	if err := db.Where("order_item_id = ?", id).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบรายการคำสั่งซื้อ"})
	}

	if err := db.Delete(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถลบรายการคำสั่งซื้อได้: " + err.Error()})
	}

	// อัปเดตยอดรวมของ Order
	if err := UpdateTotalAmount(db, orderItem.OrderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถอัปเดตยอดรวมได้: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "รายการคำสั่งซื้อถูกลบสำเร็จ"})
}

// UpdateOrderItem handles updating an order item
func UpdateOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var orderItem Models.OrderItem
	if err := db.Where("order_item_id = ?", id).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order item not found"})
	}

	type OrderItemRequest struct {
		ProductID   string  `json:"productid" validate:"required"`
		Quantity    int     `json:"quantity" validate:"required,min=1"`
		ConversRate float64 `json:"conversrate" validate:"required,min=0"`
	}

	var req OrderItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	orderItem.ProductID = req.ProductID
	orderItem.Quantity = req.Quantity
	orderItem.ConversRate = req.ConversRate

	if err := db.Save(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order item: " + err.Error()})
	}

	if err := UpdateTotalAmount(db, orderItem.OrderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update total amount: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Order item updated successfully", "data": orderItem})
}

func LookOrderItems(db *gorm.DB, c *fiber.Ctx) error {
	var orderItems []Models.OrderItem
	if err := db.Find(&orderItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find order items: " + err.Error()})
	}
	return c.JSON(fiber.Map{"data": orderItems})
}

func FindOrderItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var orderItem Models.OrderItem
	if err := db.Where("order_item_id = ?", id).First(&orderItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order item not found"})
	}
	return c.JSON(fiber.Map{"data": orderItem})
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
