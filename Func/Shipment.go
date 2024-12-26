package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddShipment(db *gorm.DB, c *fiber.Ctx) error {
	type ShipmentRequest struct {
		ShipmentNumber string `json:"shipmentnumber"`
		Status         string `json:"status"`
		FromBranchID   string `json:"frombranchid"`
		ToBranchID     string `json:"tobranchid"`
	}

	var req ShipmentRequest
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

	shipment := Models.Shipment{
		ShipmentNumber: req.ShipmentNumber,
		Status:         req.Status,
		FromBranchID:   req.FromBranchID,
		ToBranchID:     req.ToBranchID,
	}

	if err := db.Create(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New Shipment": shipment})
}

func LookShipments(db *gorm.DB, c *fiber.Ctx) error {
	var shipments []Models.Shipment
	if err := db.Find(&shipments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find shipments: " + err.Error()})
	}
	return c.JSON(fiber.Map{"Shipments": shipments})
}

func FindShipment(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipment Models.Shipment
	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}
	return c.JSON(fiber.Map{"Shipment": shipment})
}

func DeleteShipment(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipment Models.Shipment
	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}
	if err := db.Delete(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete shipment: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": "Succeed"})
}

func UpdateShipment(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipment Models.Shipment
	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}

	type ShipmentRequest struct {
		ShipmentNumber string `json:"shipmentnumber"`
		Status         string `json:"status"`
		FromBranchID   string `json:"frombranchid"`
		ToBranchID     string `json:"tobranchid"`
	}

	var req ShipmentRequest
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

	shipment.ShipmentNumber = req.ShipmentNumber
	shipment.Status = req.Status
	shipment.FromBranchID = req.FromBranchID
	shipment.ToBranchID = req.ToBranchID

	if err := db.Save(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update shipment: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
}

func ShipmentRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/Shipments", func(c *fiber.Ctx) error {
		return LookShipments(db, c)
	})
	app.Get("/Shipments/:id", func(c *fiber.Ctx) error {
		return FindShipment(db, c)
	})
	app.Post("/Shipments", func(c *fiber.Ctx) error {
		return AddShipment(db, c)
	})
	app.Put("/Shipments/:id", func(c *fiber.Ctx) error {
		return UpdateShipment(db, c)
	})
	app.Delete("/Shipments/:id", func(c *fiber.Ctx) error {
		return DeleteShipment(db, c)
	})
}
