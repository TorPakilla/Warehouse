package Func

import (
	"Api/Models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
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

type Request struct {
	RequestID    uuid.UUID `gorm:"column:request_id;primaryKey"` // Match shipment_id
	FromBranchID string    `gorm:"column:from_branch_id"`
	ToBranchID   string    `gorm:"column:to_branch_id"`
	ProductID    string    `gorm:"column:product_id"`
	Quantity     int       `gorm:"column:quantity"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (Request) TableName() string {
	return "Requests"
}

// เพิ่ม Shipment ใหม่
func AddShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	type ShipmentRequest struct {
		FromBranchID string `json:"from_branch_id" validate:"required"`
		ToBranchID   string `json:"to_branch_id" validate:"required"`
		Items        []struct {
			WarehouseInventoryID string      `json:"warehouse_inventory_id" validate:"required"`
			PosInventoryID       string      `json:"pos_inventory_id" validate:"required"`
			ProductUnitID        string      `json:"product_unit_id"`
			Quantity             json.Number `json:"quantity" validate:"required"`
		} `json:"items" validate:"required,dive"`
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format", "details": err.Error()})
	}

	if req.FromBranchID == "" || req.ToBranchID == "" || len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "FromBranchID, ToBranchID, and Items are required"})
	}

	shipmentID := uuid.New()

	if err := db.Transaction(func(tx *gorm.DB) error {
		shipment := Models.Shipment{
			ShipmentID:     shipmentID.String(),
			ShipmentNumber: GenerateULID(),
			FromBranchID:   req.FromBranchID,
			ToBranchID:     req.ToBranchID,
			Status:         "Pending",
			ShipmentDate:   time.Now(),
		}

		if err := tx.Save(&shipment).Error; err != nil {
			return err
		}

		for _, item := range req.Items {
			quantity, err := item.Quantity.Int64()
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Invalid quantity format")
			}

			productUnitID := item.ProductUnitID
			if productUnitID == "" {
				productUnitID = uuid.New().String()
			}

			var posInventory Inventory
			if err := posDB.Where("inventory_id = ?", item.PosInventoryID).First(&posInventory).Error; err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Invalid PosInventoryID")
			}

			shipmentItem := Models.ShipmentItem{
				ShipmentListID:       uuid.New().String(),
				ShipmentID:           shipmentID.String(),
				WarehouseInventoryID: item.WarehouseInventoryID,
				PosInventoryID:       item.PosInventoryID,
				ProductUnitID:        productUnitID,
				Status:               "Pending",
				Quantity:             int(quantity),
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}

			if err := tx.Create(&shipmentItem).Error; err != nil {
				return err
			}

			request := Request{
				RequestID:    shipmentID,
				FromBranchID: req.FromBranchID,
				ToBranchID:   req.ToBranchID,
				ProductID:    posInventory.ProductID.String(),
				Quantity:     int(quantity),
				Status:       "Pending",
				CreatedAt:    time.Now(),
			}

			if err := posDB.Create(&request).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction failed", "details": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Shipment created successfully", "shipment_id": shipmentID.String()})
}

// อัพเดตสถานะของ Shipment
func UpdateShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id") // รับ Shipment ID
	var shipment Models.Shipment

	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		log.Println("Error finding shipment:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}

	type ShipmentRequest struct {
		Status string `json:"status"`
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		log.Println("Error parsing request:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	allowedStatuses := map[string]bool{"Pending": true, "Approved": true, "Rejected": true}
	if !allowedStatuses[req.Status] {
		log.Println("Invalid status provided:", req.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	if req.Status == "Approved" {
		var shipmentItems []Models.ShipmentItem
		if err := db.Where("shipment_id = ?", id).Find(&shipmentItems).Error; err != nil {
			log.Println("Error retrieving shipment items:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve shipment items"})
		}

		if len(shipmentItems) == 0 {
			log.Println("No shipment items found for shipment ID:", id)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No shipment items found for this shipment"})
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, item := range shipmentItems {
				log.Printf("Processing item: %+v\n", item)

				var inventory Inventory
				if err := tx.Where("inventory_id = ?", item.WarehouseInventoryID).First(&inventory).Error; err != nil {
					log.Printf("Error finding inventory for ID %s: %v\n", item.WarehouseInventoryID, err)
					return fmt.Errorf("failed to find inventory for item: %s", item.WarehouseInventoryID)
				}

				log.Printf("Before update - Inventory ID: %s, Current Quantity: %d, Required Quantity: %d",
					item.WarehouseInventoryID, inventory.Quantity, item.Quantity)

				if err := tx.Model(&Inventory{}).
					Where("inventory_id = ?", item.WarehouseInventoryID).
					Where("quantity >= ?", item.Quantity).
					Update("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
					log.Printf("Failed to decrease Warehouse Inventory for ID %s: %v", item.WarehouseInventoryID, err)
					return fmt.Errorf("failed to update Warehouse Inventory for item: %s", item.WarehouseInventoryID)
				}

				if err := tx.Where("inventory_id = ?", item.WarehouseInventoryID).First(&inventory).Error; err != nil {
					log.Printf("Error fetching inventory after update for ID %s: %v", item.WarehouseInventoryID, err)
					return fmt.Errorf("failed to fetch inventory after update for WarehouseInventoryID %s", item.WarehouseInventoryID)
				}

				log.Printf("After update - Inventory ID: %s, New Quantity: %d", item.WarehouseInventoryID, inventory.Quantity)
			}
			return nil
		}); err != nil {
			log.Println("Error during inventory update transaction:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory", "details": err.Error()})
		}
	}

	// อัปเดตสถานะของ Shipment
	shipment.Status = req.Status
	shipment.UpdatedAt = time.Now()
	if err := db.Save(&shipment).Error; err != nil {
		log.Println("Error updating shipment:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update shipment"})
	}

	log.Println("Shipment updated successfully:", shipment)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Shipment updated successfully",
		"shipment": shipment,
	})
}

func AutoUpdateShipments(db *gorm.DB) error {
	var shipments []Models.Shipment

	if err := db.Where("status = ?", "Approved").Find(&shipments).Error; err != nil {
		return err
	}

	for _, shipment := range shipments {
		var shipmentItems []Models.ShipmentItem
		if err := db.Where("shipment_id = ?", shipment.ShipmentID).Find(&shipmentItems).Error; err != nil {
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, item := range shipmentItems {
				if err := tx.Model(&Inventory{}).
					Where("inventory_id = ?", item.WarehouseInventoryID).
					Where("quantity >= ?", item.Quantity).
					Update("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
					return fmt.Errorf("not enough inventory for WarehouseInventoryID %s", item.WarehouseInventoryID)
				}
			}

			shipment.Status = "Completed"
			shipment.UpdatedAt = time.Now()
			if err := tx.Save(&shipment).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			continue
		}
	}

	return nil
}

var notFoundRequests = make(map[uuid.UUID]bool)

func SyncRequestStatusWithWarehouse(db *gorm.DB, posDB *gorm.DB) error {
	var requests []Request

	// ดึงข้อมูล Requests ที่มีสถานะเปลี่ยนแปลง (complete หรือ reject)
	if err := posDB.Where("status IN ?", []string{"complete", "reject"}).Find(&requests).Error; err != nil {
		return err
	}

	for _, request := range requests {
		// ข้าม RequestID ที่เคยหาไม่เจอในรอบก่อนหน้า
		if notFoundRequests[request.RequestID] {
			continue
		}

		var shipment Models.Shipment

		if err := db.Where("shipment_id = ?", request.RequestID).First(&shipment).Error; err != nil {
			// เก็บ RequestID ที่หาไม่เจอเข้าไปในแคช
			notFoundRequests[request.RequestID] = true
			continue
		}

		// ดำเนินการอัปเดตตามปกติ
		if err := db.Transaction(func(tx *gorm.DB) error {
			switch request.Status {
			case "complete":
				shipment.Status = "Approved"
				if err := tx.Save(&shipment).Error; err != nil {
					return fmt.Errorf("failed to update shipment status: %v", err)
				}

				if err := tx.Model(&Inventory{}).
					Where("branch_id = ? AND product_id = ?", request.FromBranchID, request.ProductID).
					Update("quantity", gorm.Expr("quantity - ?", request.Quantity)).Error; err != nil {
					return fmt.Errorf("failed to update source inventory: %v", err)
				}

				if err := tx.Model(&Inventory{}).
					Where("branch_id = ? AND product_id = ?", request.ToBranchID, request.ProductID).
					Update("quantity", gorm.Expr("quantity + ?", request.Quantity)).Error; err != nil {
					return fmt.Errorf("failed to update destination inventory: %v", err)
				}

			case "reject":
				shipment.Status = "Rejected"
				if err := tx.Save(&shipment).Error; err != nil {
					return fmt.Errorf("failed to update shipment status: %v", err)
				}
			}

			if err := posDB.Save(&request).Error; err != nil {
				return fmt.Errorf("failed to update request status in POS: %v", err)
			}

			return nil
		}); err != nil {
			log.Printf("Error syncing request %s with warehouse: %v\n", request.RequestID, err)
			continue
		}

		// ลบ RequestID ออกจากแคช หากดำเนินการสำเร็จ
		delete(notFoundRequests, request.RequestID)
	}

	return nil
}

func StartSyncScheduler(db *gorm.DB, posDB *gorm.DB) {
	scheduler := gocron.NewScheduler(time.UTC)

	scheduler.Every(10).Seconds().Do(func() {
		_ = SyncRequestStatusWithWarehouse(db, posDB)
	})

	scheduler.Every(10).Seconds().Do(func() {
		_ = AutoUpdateShipments(db)
	})

	scheduler.StartAsync()
}

func LookShipments(db *gorm.DB, c *fiber.Ctx) error {
	var shipments []Models.Shipment
	if err := db.Find(&shipments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch shipments",
		})
	}

	return c.JSON(fiber.Map{
		"Shipments": shipments,
	})
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
