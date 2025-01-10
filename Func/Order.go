package Func

import (
	"Api/Models"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

func GenerateULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func AddOrder(db *gorm.DB, c *fiber.Ctx) error {
	type OrderRequest struct {
		Status     string `json:"status"`
		SupplierID string `json:"supplierid"`
		EmployeeID string `json:"employeeid"`
	}

	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"supplierid": true,
		"employeeid": true,
		"status":     true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	req.Status = "Pending"

	CheckStatus := map[string]bool{
		"Pending": true,
	}

	if !CheckStatus[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed "})
	}

	order := Models.Order{
		Status:     req.Status,
		SupplierID: req.SupplierID,
		EmployeeID: req.EmployeeID,
	}

	order.OrderNumber = GenerateULID()

	if err := db.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New": order})
}

func UpdateTotalAmount(db *gorm.DB, orderID string) error {
	var orderItems []Models.OrderItem
	if err := db.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		return err
	}

	var totalAmount float64
	for _, item := range orderItems {
		totalAmount += float64(item.Quantity) * item.UnitPrice
	}

	if err := db.Model(&Models.Order{}).Where("order_id = ?", orderID).Update("total_amount", totalAmount).Error; err != nil {
		return err
	}

	return nil
}

func LookOrders(db *gorm.DB, c *fiber.Ctx) error {
	var orders []Models.Order
	if err := db.Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find orders: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "Order", "Data": orders})
}

func FindOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var order Models.Order
	if err := db.Where("order_id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}
	return c.JSON(fiber.Map{"This": "Order", "Data": order})
}

func DeleteOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var order Models.Order
	if err := db.Where("order_id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}
	if err := db.Delete(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete order: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": "Succeed"})
}

func fetchConversionRate(productUnitID string, db *gorm.DB) (int, error) {
	var productUnit Models.ProductUnit
	if err := db.Where("product_unit_id = ?", productUnitID).First(&productUnit).Error; err != nil {
		return 0, err
	}
	if productUnit.ConversRate == nil {
		return 0, fmt.Errorf("ไม่มีค่าการแปลงสำหรับ Product Unit ID %s", productUnitID)
	}
	return *productUnit.ConversRate, nil
}

func UpdateOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var order Models.Order
	if err := db.Where("order_id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	if order.Status != "Pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only orders with Pending status can be updated"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"supplierid": true,
		"employeeid": true,
		"status":     true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	var req struct {
		Status string `json:"status"`
	}

	if status, ok := body["status"].(string); ok && status != "" {
		req.Status = status
	}

	validStatuses := map[string]bool{
		"Pending":  true,
		"Approved": true,
		"Rejected": true,
	}

	if req.Status != "" && !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status. Allowed statuses: Pending, Approved, Rejected"})
	}

	if req.Status == "Approved" {
		if err := db.Transaction(func(tx *gorm.DB) error {
			var orderItems []Models.OrderItem
			if err := tx.Where("order_id = ?", id).Find(&orderItems).Error; err != nil {
				return fmt.Errorf("failed to fetch order items: %w", err)
			}

			for _, item := range orderItems {
				conversionRate, err := fetchConversionRate(item.ProductUnitID, tx)
				if err != nil {
					return fmt.Errorf("failed to fetch conversion rate for product unit ID %s: %w", item.ProductUnitID, err)
				}

				additionalValue := item.Quantity * conversionRate
				if err := tx.Model(&Models.Inventory{}).
					Where("product_unit_id = ?", item.ProductUnitID).
					UpdateColumn("quantity", gorm.Expr("quantity + ?", additionalValue)).Error; err != nil {
					return fmt.Errorf("failed to update inventory for product unit ID %s: %w", item.ProductUnitID, err)
				}
			}

			order.Status = req.Status
			order.UpdateAt = time.Now()
			if err := tx.Save(&order).Error; err != nil {
				return fmt.Errorf("failed to update order status: %w", err)
			}

			return nil
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order: " + err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order updated successfully",
	})
}

func OrderRoutes(app *fiber.App, db *gorm.DB) {
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
