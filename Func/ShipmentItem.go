package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	type ShipmentItemRequest struct {
		ShipmentID    string `json:"shipmentid"`
		ProductUnitID string `json:"productunitid"`
		Quantity      int    `json:"quantity"`
	}

	var req ShipmentItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	shipmentItem := Models.ShipmentItem{
		ShipmentID:    req.ShipmentID,
		ProductUnitID: req.ProductUnitID,
		Quantity:      req.Quantity,
	}

	if err := db.Create(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New Shipment Item": shipmentItem})
}

func LookShipmentItems(db *gorm.DB, c *fiber.Ctx) error {
	var shipmentItems []Models.ShipmentItem
	if err := db.Find(&shipmentItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find shipment items: " + err.Error()})
	}
	return c.JSON(fiber.Map{"Shipment Items": shipmentItems})
}

func FindShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipmentItem Models.ShipmentItem
	if err := db.Where("shipment_list_id = ?", id).First(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment item not found"})
	}
	return c.JSON(fiber.Map{"Shipment Item": shipmentItem})
}

func DeleteShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipmentItem Models.ShipmentItem
	if err := db.Where("shipment_list_id = ?", id).First(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment item not found"})
	}
	if err := db.Delete(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete shipment item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": "Succeed"})
}

func UpdateShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipmentItem Models.ShipmentItem
	if err := db.Where("shipment_list_id = ?", id).First(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment item not found"})
	}

	type ShipmentItemRequest struct {
		ShipmentID    string `json:"shipmentid"`
		ProductUnitID string `json:"productunitid"`
		Quantity      int    `json:"quantity"`
	}

	var req ShipmentItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	shipmentItem.ShipmentID = req.ShipmentID
	shipmentItem.ProductUnitID = req.ProductUnitID
	shipmentItem.Quantity = req.Quantity

	if err := db.Save(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update shipment item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
}

func ShipmentItemRoutes(app *fiber.App, db *gorm.DB) {
	app.Use(func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "God" && role != "Manager" && role != "Stock" {
			return c.Next()
		}

		if role != "Account" {
			if c.Method() != "GET" && c.Method() != "UPDATE" {
				return c.Next()
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
			}
		}

		if role != "Audit" {
			if c.Method() != "GET" {
				return c.Next()
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	})

	app.Get("/ShipmentItems", func(c *fiber.Ctx) error {
		return LookShipmentItems(db, c)
	})

	app.Get("/ShipmentItems/:id", func(c *fiber.Ctx) error {
		return FindShipmentItem(db, c)
	})

	app.Post("/ShipmentItems", func(c *fiber.Ctx) error {
		return AddShipmentItem(db, c)
	})

	app.Put("/ShipmentItems/:id", func(c *fiber.Ctx) error {
		return UpdateShipmentItem(db, c)
	})

	app.Delete("/ShipmentItems/:id", func(c *fiber.Ctx) error {
		return DeleteShipmentItem(db, c)
	})
}
