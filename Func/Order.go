package Func

import (
	"Api/Models"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type OrderItemRequest struct {
	ProductID string  `json:"productid" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	UnitPrice float64 `json:"unitprice" validate:"required,min=0"`
}

func GenerateULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func parseUUIDPointer(id *string) *uuid.UUID {
	if id == nil {
		return nil
	}
	u, err := uuid.Parse(*id)
	if err != nil {
		return nil
	}
	return &u
}

func AddOrder(db *gorm.DB, c *fiber.Ctx) error {
	type OrderRequest struct {
		SupplierID  string             `json:"supplier_id" validate:"required"`
		EmployeesID *string            `json:"employees_id"`
		OrderItems  []OrderItemRequest `json:"order_items" validate:"required"`
	}

	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// Verify Supplier
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", req.SupplierID).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}

	// Verify EmployeesID if provided
	if req.EmployeesID != nil {
		var employee Models.Employees
		if err := db.Where("employees_id = ?", *req.EmployeesID).First(&employee).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
		}
	}

	// Create Order
	order := Models.Order{
		OrderID:     uuid.New().String(),
		OrderNumber: GenerateULID(),
		Status:      "Pending", // Always set to Pending
		SupplierID:  uuid.MustParse(req.SupplierID),
		EmployeesID: parseUUIDPointer(req.EmployeesID),
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order"})
	}

	// Create OrderItems
	for _, item := range req.OrderItems {
		var product Models.Product
		if err := db.Where("product_id = ?", item.ProductID).First(&product).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found: " + item.ProductID})
		}

		// Verify convers_rate for the product
		var productUnit Models.ProductUnit
		if err := db.Where("product_id = ?", item.ProductID).First(&productUnit).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ProductUnit not found for product: " + item.ProductID})
		}

		if productUnit.ConversRate <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid convers_rate for product: " + item.ProductID})
		}

		// Calculate final quantity using convers_rate
		finalQuantity := item.Quantity * productUnit.ConversRate
		if finalQuantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Final quantity cannot be zero or negative for product: " + item.ProductID})
		}

		orderItem := Models.OrderItem{
			OrderItemID: uuid.New().String(),
			OrderID:     order.OrderID,
			ProductID:   item.ProductID,
			Quantity:    finalQuantity,
			ConversRate: float64(productUnit.ConversRate),
			CreatedAt:   time.Now(),
		}

		if err := db.Create(&orderItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order item"})
		}
	}

	// Update total amount
	if err := UpdateTotalAmount(db, order.OrderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order total amount"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Order created successfully", "data": order})
}

func UpdateTotalAmount(db *gorm.DB, orderID string) error {
	var orderItems []Models.OrderItem
	if err := db.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		return err
	}

	var totalAmount float64
	for _, item := range orderItems {
		totalAmount += float64(item.Quantity) * item.ConversRate
	}

	return db.Model(&Models.Order{}).Where("order_id = ?", orderID).Update("total_amount", totalAmount).Error
}

func UpdateOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")

	// ตรวจสอบว่า Order มีอยู่ในระบบหรือไม่
	var order Models.Order
	if err := db.Where("order_id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	// อนุญาตให้อัปเดตเฉพาะ Order ที่อยู่ในสถานะ Pending เท่านั้น
	if order.Status != "Pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only orders with Pending status can be updated"})
	}

	// รับค่าที่ส่งมาสำหรับอัปเดต
	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// ตรวจสอบว่าสถานะใหม่เป็นค่าที่ถูกต้อง
	validStatuses := map[string]bool{"Pending": true, "Approved": true, "Rejected": true}
	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	// ถ้าสถานะใหม่เป็น Approved ให้ปรับปรุงคลังสินค้า (Inventory)
	if req.Status == "Approved" {
		var orderItems []Models.OrderItem
		if err := db.Where("order_id = ?", id).Find(&orderItems).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch order items"})
		}

		// Debug: แสดงรายการ OrderItems
		fmt.Printf("Order Items for Order ID: %s\n", id)
		for _, item := range orderItems {
			fmt.Printf("Product ID: %s, Quantity: %d\n", item.ProductID, item.Quantity)

			// อัปเดต Inventory
			if err := db.Model(&Models.Inventory{}).
				Where("product_id = ?", item.ProductID).
				UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity)).Error; err != nil {
				fmt.Printf("Failed to update inventory for Product ID: %s\n", item.ProductID)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory for product: " + item.ProductID})
			}
		}
	}

	// อัปเดตสถานะของ Order
	order.Status = req.Status
	order.UpdatedAt = time.Now()
	if err := db.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Order updated successfully"})
}

func LookOrders(db *gorm.DB, c *fiber.Ctx) error {
	var orders []Models.Order
	if err := db.Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}
	return c.JSON(fiber.Map{"data": orders})
}

func FindOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var order Models.Order
	if err := db.Preload("Supplier").Where("order_id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}
	return c.JSON(fiber.Map{"data": order})
}

func DeleteOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Where("order_id = ?", id).Delete(&Models.Order{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete order"})
	}
	return c.JSON(fiber.Map{"message": "Order deleted successfully"})
}

func OrderRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/Orders", func(c *fiber.Ctx) error {
		return LookOrders(db, c)
	})

	app.Get("/Orders/:id", func(c *fiber.Ctx) error {
		return FindOrder(db, c)
	})

	app.Post("/Orders", func(c *fiber.Ctx) error {
		return AddOrder(db, c)
	})

	app.Put("/Orders/:id", func(c *fiber.Ctx) error {
		return UpdateOrder(db, c)
	})

	app.Delete("/Orders/:id", func(c *fiber.Ctx) error {
		return DeleteOrder(db, c)
	})
}
