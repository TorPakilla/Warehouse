package Func

import (
	"Api/Models"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// เพิ่มข้อมูล ShipmentItem
func AddShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	type ShipmentItemRequest struct {
		ShipmentID           string `json:"shipmentID" validate:"required"`
		ProductID            string `json:"productID" validate:"required"`
		WarehouseInventoryID string `json:"warehouseInventoryID" validate:"required"`
		POSInventoryID       string `json:"posInventoryID" validate:"required"`
		Quantity             int    `json:"quantity" validate:"required,min=1"`
		Status               string `json:"status" validate:"required"`
		FromBranchID         string `json:"fromBranchID" validate:"required"`
		ToBranchID           string `json:"toBranchID" validate:"required"`
	}

	var req ShipmentItemRequest
	if err := c.BodyParser(&req); err != nil {
		log.Println("Error parsing request:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	log.Println("Received Request:", req)

	var existingShipment Models.Shipment
	err := db.Where("shipment_id = ?", req.ShipmentID).First(&existingShipment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("Shipment not found, creating new shipment:", req.ShipmentID)

			newShipment := Models.Shipment{
				ShipmentID:     req.ShipmentID, // ใช้ ShipmentID เดิม
				ShipmentNumber: fmt.Sprintf("SH-%d", time.Now().UnixNano()),
				FromBranchID:   req.FromBranchID,
				ToBranchID:     req.ToBranchID,
				Status:         "Pending",
				ShipmentDate:   time.Now(),
			}

			if err := db.Create(&newShipment).Error; err != nil {
				log.Println("Error creating new shipment:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment"})
			}

			existingShipment = newShipment
		} else {
			log.Println("Error checking existing shipment:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking existing shipment"})
		}
	}

	shipmentItem := Models.ShipmentItem{
		ShipmentID:           existingShipment.ShipmentID,
		ProductUnitID:        req.ProductID,
		WarehouseInventoryID: req.WarehouseInventoryID,
		PosInventoryID:       req.POSInventoryID,
		Quantity:             req.Quantity,
		Status:               req.Status,
	}

	if err := db.Create(&shipmentItem).Error; err != nil {
		log.Println("Error creating shipment item:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment item"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Shipment and item added successfully", "data": shipmentItem})
}

// ดูข้อมูล ShipmentItem ทั้งหมด
func LookShipmentItems(db *gorm.DB, c *fiber.Ctx) error {
	var shipmentItems []Models.ShipmentItem
	if err := db.Find(&shipmentItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find shipment items: " + err.Error()})
	}
	return c.JSON(fiber.Map{"data": shipmentItems})
}

// หาข้อมูล ShipmentItem ตาม ID
func FindShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipmentItem Models.ShipmentItem
	if err := db.Where("shipment_list_id = ?", id).First(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment item not found"})
	}
	return c.JSON(fiber.Map{"data": shipmentItem})
}

// ลบข้อมูล ShipmentItem ตาม ID
func DeleteShipmentItem(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipmentItem Models.ShipmentItem
	if err := db.Where("shipment_list_id = ?", id).First(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment item not found"})
	}
	if err := db.Delete(&shipmentItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete shipment item: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Shipment item deleted successfully"})
}

// อัพเดทข้อมูล ShipmentItem
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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Shipment item updated successfully", "data": shipmentItem})
}

func ShipmentItemRoutes(app *fiber.App, db *gorm.DB) {
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
