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

// AddShipment handles the creation of a shipment and manages inventory movement
func AddShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	type ShipmentRequest struct {
		FromBranchID uuid.UUID             `json:"fromBranchId" validate:"required"`
		ToBranchID   uuid.UUID             `json:"toBranchId" validate:"required"`
		Status       string                `json:"status" validate:"required,oneof=Pending Approved Rejected"`
		Items        []Models.ShipmentItem `json:"items" validate:"required,dive"`
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format", "details": err.Error()})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Shipment items are required"})
	}

	// Check if ToBranch exists in POS
	var toBranch Inventory
	if err := posDB.Where("branch_id = ?", req.ToBranchID).First(&toBranch).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ToBranch is not a valid POS branch"})
	}

	// Create Shipment
	shipment := Models.Shipment{
		ShipmentID:   uuid.New().String(),
		FromBranchID: req.FromBranchID.String(), // แปลง UUID เป็น String
		ToBranchID:   req.ToBranchID.String(),   // แปลง UUID เป็น String
		Status:       req.Status,
		ShipmentDate: time.Now(),
	}

	if err := db.Create(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shipment"})
	}

	for _, item := range req.Items {
		// ตรวจสอบ Inventory ของ Warehouse
		var warehouseInventory Inventory
		if err := db.Where("inventory_id = ?", item.InventoryID).First(&warehouseInventory).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Inventory not found in Warehouse"})
		}

		if warehouseInventory.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot add items with 0 quantity in inventory"})
		}

		if warehouseInventory.Quantity < item.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough inventory in Warehouse"})
		}

		// ลดจำนวนสินค้าใน Warehouse
		warehouseInventory.Quantity -= item.Quantity
		if err := db.Save(&warehouseInventory).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update Warehouse inventory"})
		}

		// เพิ่มหรืออัปเดต Inventory ใน POS
		var posInventory Inventory
		err := posDB.Where("product_id = ? AND branch_id = ?", warehouseInventory.ProductID, req.ToBranchID).
			First(&posInventory).Error
		if err == gorm.ErrRecordNotFound {
			posInventory = Inventory{
				InventoryID: uuid.New(),
				ProductID:   warehouseInventory.ProductID,
				BranchID:    req.ToBranchID,
				Quantity:    item.Quantity,
				UpdatedAt:   time.Now(),
			}
			if err := posDB.Create(&posInventory).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create POS inventory"})
			}
		} else if err == nil {
			posInventory.Quantity += item.Quantity
			if err := posDB.Save(&posInventory).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update POS inventory"})
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Shipment created successfully", "shipment": shipment})
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

	if shipment.Status != "Pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only shipments with Pending status can be updated"})
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"status":       true,
		"frombranchid": true,
		"tobranchid":   true,
		"inventory_id": true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	type ShipmentRequest struct {
		Status      string `json:"status"`
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

	if req.Status == "Approved" {
		var inventoryItem Inventory
		if err := posDB.Where("inventory_id = ?", req.InventoryID).First(&inventoryItem).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory item not found"})
		}

		targetBranchID, err := uuid.Parse(shipment.ToBranchID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ToBranchID format"})
		}

		if inventoryItem.BranchID != targetBranchID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Inventory does not belong to the target branch"})
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
				posItem.UpdatedAt = time.Now()
				inventoryPosList = append(inventoryPosList, posItem)
			}

			for _, item := range inventoryPosList {
				if err := posDB.Model(&Inventory{}).
					Where("inventory_id = ?", item.InventoryID).
					UpdateColumn("quantity", item.Quantity).Error; err != nil {
					return err
				}

				if err := posDB.Model(&Inventory{}).
					Where("inventory_id = ?", item.InventoryID).
					UpdateColumn("updated_at", item.UpdatedAt).Error; err != nil {
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
		return AddShipment(db, posDB, c)
	})

	app.Put("/Shipments/:id", func(c *fiber.Ctx) error {
		return UpdateShipment(db, posDB, c)
	})

	app.Delete("/Shipments/:id", func(c *fiber.Ctx) error {
		return DeleteShipment(db, c)
	})
}
