package Func

import (
	"Api/Models"
	"crypto/rand"
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

func UpdateOrder(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var order Models.Order
	if err := db.Where("order_id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	type OrderRequest struct {
		Status string `json:"status"`
	}

	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	CheckStatus := map[string]bool{
		"Pending":  true,
		"Approved": true,
		"Rejected": true,
	}

	if !CheckStatus[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status. Allowed statuses: Pending, Approved, Rejected"})
	}

	if order.Status == "Approved" || order.Status == "Rejected" {
		if req.Status == "Pending" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot revert status to Pending"})
		}
	}

	if err := db.Transaction(func(db *gorm.DB) error {
		if req.Status == "Approved" {
			var orderItems []Models.OrderItem
			if err := db.Where("order_id = ?", id).Find(&orderItems).Error; err != nil {
				return err
			}

			for _, item := range orderItems {
				if err := db.Model(&Models.Inventory{}).
					Where("product_unit_id = ?", item.ProductUnitID).
					UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity)).Error; err != nil {
					return err
				}
			}
		}

		order.UpdateAt = time.Now()
		order.Status = req.Status
		if err := db.Save(&order).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order item: " + err.Error()})
		}

		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Order updated successfully"})
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
