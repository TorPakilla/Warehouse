package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddOrder(db *gorm.DB, c *fiber.Ctx) error {
	type OrderRequest struct {
		OrderNumber string  `json:"ordernumber"`
		TotalAmount float64 `json:"totalamount"`
		Status      string  `json:"status"`
		SupplierID  string  `json:"supplierid"`
		EmployeeID  string  `json:"employeeid"`
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed "})
	}

	order := Models.Order{
		OrderNumber: req.OrderNumber,
		TotalAmount: req.TotalAmount,
		Status:      req.Status,
		SupplierID:  req.SupplierID,
		EmployeeID:  req.EmployeeID,
	}

	if err := db.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New": order})
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
		OrderNumber string  `json:"ordernumber"`
		TotalAmount float64 `json:"totalamount"`
		Status      string  `json:"status"`
		SupplierID  string  `json:"supplierid"`
		EmployeeID  string  `json:"employeeid"`
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed "})
	}

	order.OrderNumber = req.OrderNumber
	order.TotalAmount = req.TotalAmount
	order.Status = req.Status
	order.SupplierID = req.SupplierID
	order.EmployeeID = req.EmployeeID

	if err := db.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
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
