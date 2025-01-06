package Func

import (
	"Api/Models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Inventory struct {
	InventoryID uuid.UUID `gorm:"column:inventory_id;primaryKey"`
	ProductID   uuid.UUID `gorm:"column:product_id"`
	BranchID    uuid.UUID `gorm:"column:branch_id"`
	Quantity    int       `gorm:"column:quantity"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Inventory) TableName() string {
	return "Inventory"
}

func AddShipment(db *gorm.DB, c *fiber.Ctx) error {
	type ShipmentRequest struct {
		Status       string `json:"status"`
		FromBranchID string `json:"frombranchid"`
		ToBranchID   string `json:"tobranchid"`
	}

	var req ShipmentRequest
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

	shipment := Models.Shipment{
		Status:       req.Status,
		FromBranchID: req.FromBranchID,
		ToBranchID:   req.ToBranchID,
	}

	shipment.ShipmentNumber = GenerateULID()

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

func UpdateShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipment Models.Shipment
	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}

	type ShipmentRequest struct {
		Status      string `json:"status"`
		ProductID   string `json:"product_id"`
		InventoryID string `json:"inventory_id"`
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedStatuses := map[string]bool{
		"Pending":  true,
		"Approved": true,
		"Rejected": true,
	}

	if !allowedStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status. Allowed: Pending, Approved, Rejected"})
	}

	if shipment.Status == "Approved" || shipment.Status == "Rejected" {
		if req.Status == "Pending" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot revert status to Pending"})
		}
	}

	if err := db.Transaction(func(db *gorm.DB) error {
		if req.Status == "Approved" {
			var shipmentItems []Models.ShipmentItem
			if err := db.Where("shipment_id = ?", id).Find(&shipmentItems).Error; err != nil {
				return err
			}

			for _, item := range shipmentItems {
				if err := db.Model(&Models.Inventory{}).
					Where("product_unit_id = ?", item.ProductUnitID).
					UpdateColumn("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
					return err
				}
			}
			var inventoryPosList []Inventory
			for _, item := range shipmentItems {
				var posItem Inventory
				if err := posDB.Where("inventory_id = ?", req.InventoryID).First(&posItem).Error; err != nil {
					return err
				}

				posItem.Quantity += item.Quantity
				inventoryPosList = append(inventoryPosList, posItem)
			}

			for _, item := range inventoryPosList {
				if err := posDB.Model(&Inventory{}).
					Where("inventory_id = ?", item.InventoryID).
					UpdateColumn("quantity", item.Quantity).Error; err != nil {
					return err
				}
			}
		}

		shipment.Status = req.Status
		shipment.UpdateAt = time.Now()

		if err := db.Save(&shipment).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update shipment: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
}

func ShipmentRoutes(app *fiber.App, db *gorm.DB, posDB *gorm.DB) {
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
		return UpdateShipment(db, posDB, c)
	})

	app.Delete("/Shipments/:id", func(c *fiber.Ctx) error {
		return DeleteShipment(db, c)
	})
}
