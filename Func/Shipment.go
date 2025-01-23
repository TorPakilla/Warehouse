package Func

import (
	"Api/Models"
	"encoding/json"
	"fmt"
	"log"
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

// AddShipment handles adding a new shipment along with requests
func AddShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	type ShipmentRequest struct {
		FromBranchID string `json:"from_branch_id" validate:"required"`
		ToBranchID   string `json:"to_branch_id" validate:"required"`
		Items        []struct {
			WarehouseInventoryID string      `json:"warehouse_inventory_id" validate:"required"`
			PosInventoryID       string      `json:"pos_inventory_id" validate:"required"`
			ProductUnitID        string      `json:"product_unit_id"` // optional
			Quantity             json.Number `json:"quantity" validate:"required"`
		} `json:"items" validate:"required,dive"`
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format", "details": err.Error()})
	}

	// Check required fields
	if req.FromBranchID == "" || req.ToBranchID == "" || len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "FromBranchID, ToBranchID, and Items are required"})
	}

	// Generate ShipmentID once
	shipmentID := uuid.New()
	log.Printf("Generated ShipmentID: %s\n", shipmentID.String())

	// Start transaction
	if err := db.Transaction(func(tx *gorm.DB) error {
		// Create Shipment
		shipment := Models.Shipment{
			ShipmentID:     shipmentID.String(),
			ShipmentNumber: GenerateULID(),
			FromBranchID:   req.FromBranchID,
			ToBranchID:     req.ToBranchID,
			Status:         "Pending",
			ShipmentDate:   time.Now(),
		}

		if err := tx.Create(&shipment).Error; err != nil {
			log.Println("Error creating shipment:", err)
			return err
		}

		// Create ShipmentItems and Requests
		for _, item := range req.Items {
			quantity, err := item.Quantity.Int64()
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Invalid quantity format")
			}

			productUnitID := item.ProductUnitID
			if productUnitID == "" {
				productUnitID = uuid.New().String()
			}

			// Validate PosInventory
			var posInventory Inventory
			if err := posDB.Where("inventory_id = ?", item.PosInventoryID).First(&posInventory).Error; err != nil {
				log.Println("Error finding PosInventory:", err)
				return fiber.NewError(fiber.StatusBadRequest, "Invalid PosInventoryID")
			}

			// Create ShipmentItem
			shipmentItem := Models.ShipmentItem{
				ShipmentListID:       uuid.New().String(),
				ShipmentID:           shipmentID.String(), // Use same ShipmentID
				WarehouseInventoryID: item.WarehouseInventoryID,
				PosInventoryID:       item.PosInventoryID,
				ProductUnitID:        productUnitID,
				Status:               "Pending",
				Quantity:             int(quantity),
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}

			if err := tx.Create(&shipmentItem).Error; err != nil {
				log.Println("Error creating shipment item:", err)
				return err
			}

			// Create Request
			request := Request{
				RequestID:    shipmentID, // Use same ShipmentID
				FromBranchID: req.FromBranchID,
				ToBranchID:   req.ToBranchID,
				ProductID:    posInventory.ProductID.String(),
				Quantity:     int(quantity),
				Status:       "pending",
				CreatedAt:    time.Now(),
			}

			if err := posDB.Create(&request).Error; err != nil {
				log.Println("Error creating request in posDB:", err)
				return err
			}
		}

		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction failed", "details": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Shipment created successfully", "shipment_id": shipmentID.String()})
}

// UpdateShipment updates the status of a shipment
// UpdateShipment updates the status of a shipment
func UpdateShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shipment Models.Shipment

	// Fetch shipment
	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		log.Println("Error finding shipment:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}

	type ShipmentRequest struct {
		Status string `json:"status"`
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		log.Println("Error parsing shipment update request:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format", "details": err.Error()})
	}

	allowedStatuses := map[string]bool{"Pending": true, "Approved": true, "Rejected": true}
	if !allowedStatuses[req.Status] {
		log.Println("Invalid status provided:", req.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	// Handle "Approved" status
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

		// Start transaction for inventory updates
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, item := range shipmentItems {
				log.Printf("Processing item: %+v\n", item)

				// Decrease quantity from Warehouse Inventory
				if err := tx.Model(&Inventory{}).
					Where("inventory_id = ?", item.WarehouseInventoryID).
					Where("quantity >= ?", item.Quantity).
					Update("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
					log.Printf("Failed to update warehouse inventory: %v\n", err)
					return fmt.Errorf("failed to update Warehouse inventory for item: %s", item.WarehouseInventoryID)
				}

				// Increase quantity in POS Inventory
				if err := posDB.Model(&Inventory{}).
					Where("inventory_id = ?", item.PosInventoryID).
					Update("quantity", gorm.Expr("quantity + ?", item.Quantity)).Error; err != nil {
					log.Printf("Failed to update POS inventory: %v\n", err)
					return fmt.Errorf("failed to update POS inventory for item: %s", item.PosInventoryID)
				}
			}
			return nil
		}); err != nil {
			log.Println("Error during inventory update transaction:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory", "details": err.Error()})
		}
	}

	// Update shipment status
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

func UpdateRequestStatus(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var request Request

	// Fetch the request by ID
	if err := posDB.Where("request_id = ?", id).First(&request).Error; err != nil {
		log.Println("Error finding request:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Request not found"})
	}

	type RequestUpdate struct {
		Status string `json:"status"`
	}

	var req RequestUpdate
	if err := c.BodyParser(&req); err != nil {
		log.Println("Error parsing request update:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format", "details": err.Error()})
	}

	allowedStatuses := map[string]bool{"Pending": true, "Approved": true, "Rejected": true, "complete": true, "reject": true}
	if !allowedStatuses[req.Status] {
		log.Println("Invalid status provided:", req.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	// Start a transaction for updating both databases
	if err := db.Transaction(func(tx *gorm.DB) error {
		// Update status in posDB
		request.Status = req.Status
		if err := posDB.Save(&request).Error; err != nil {
			log.Println("Error updating request status in posDB:", err)
			return err
		}

		// Handle inventory updates if Approved or complete
		if req.Status == "Approved" || req.Status == "complete" {
			// Fetch inventory item from warehouse DB
			var warehouseInventory Inventory
			if err := tx.Where("inventory_id = ?", request.ProductID).First(&warehouseInventory).Error; err != nil {
				log.Println("Error finding inventory in warehouse DB:", err)
				return fmt.Errorf("Failed to retrieve inventory: %v", err)
			}

			// Start inventory update in both branches
			if err := tx.Transaction(func(innerTx *gorm.DB) error {
				// Decrease quantity from source branch in Warehouse
				if err := innerTx.Model(&Inventory{}).
					Where("branch_id = ?", request.FromBranchID).
					Where("quantity >= ?", request.Quantity).
					Update("quantity", gorm.Expr("quantity - ?", request.Quantity)).Error; err != nil {
					log.Println("Error updating source inventory in warehouse:", err)
					return fmt.Errorf("failed to update source inventory in warehouse: %v", err)
				}

				// Increase quantity in destination branch in Warehouse
				if err := innerTx.Model(&Inventory{}).
					Where("branch_id = ?", request.ToBranchID).
					Update("quantity", gorm.Expr("quantity + ?", request.Quantity)).Error; err != nil {
					log.Println("Error updating destination inventory in warehouse:", err)
					return fmt.Errorf("failed to update destination inventory in warehouse: %v", err)
				}

				// Sync changes with POS branch inventory
				if err := posDB.Model(&Inventory{}).
					Where("branch_id = ?", request.ToBranchID).
					Update("quantity", gorm.Expr("quantity + ?", request.Quantity)).Error; err != nil {
					log.Println("Error updating POS inventory:", err)
					return fmt.Errorf("failed to sync POS inventory: %v", err)
				}

				return nil
			}); err != nil {
				return err
			}
		} else if req.Status == "Rejected" || req.Status == "reject" {
			log.Printf("Request ID %s rejected, no inventory changes made.\n", id)
		}

		return nil
	}); err != nil {
		log.Println("Error during transaction:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update request status", "details": err.Error()})
	}

	log.Printf("Request %s updated successfully with status: %s\n", id, req.Status)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Request updated successfully",
		"request": request,
	})
}

func SyncRequestStatusWithWarehouse(db *gorm.DB, posDB *gorm.DB) error {
	var requests []Request

	// ดึงข้อมูล Requests ที่เปลี่ยนสถานะใน POS
	if err := posDB.Where("status IN ?", []string{"complete", "reject"}).Find(&requests).Error; err != nil {
		log.Println("Error fetching requests from POS:", err)
		return err
	}

	for _, request := range requests {
		var shipment Models.Shipment

		// ค้นหา shipment ใน Warehouse
		if err := db.Where("shipment_id = ?", request.RequestID).First(&shipment).Error; err != nil {
			log.Printf("Shipment not found for Request ID %s: %v\n", request.RequestID, err)
			continue
		}

		// เริ่มอัปเดตสถานะ
		if err := db.Transaction(func(tx *gorm.DB) error {
			if request.Status == "complete" {
				// อัปเดตสถานะ Shipment
				shipment.Status = "Approved"
				if err := tx.Save(&shipment).Error; err != nil {
					return fmt.Errorf("failed to update shipment status: %v", err)
				}

				// อัปเดต Inventory
				if err := tx.Model(&Inventory{}).
					Where("branch_id = ? AND product_id = ?", request.FromBranchID, request.ProductID).
					Update("quantity", gorm.Expr("quantity - ?", request.Quantity)).Error; err != nil {
					return fmt.Errorf("failed to update warehouse inventory: %v", err)
				}

				if err := tx.Model(&Inventory{}).
					Where("branch_id = ? AND product_id = ?", request.ToBranchID, request.ProductID).
					Update("quantity", gorm.Expr("quantity + ?", request.Quantity)).Error; err != nil {
					return fmt.Errorf("failed to update warehouse inventory: %v", err)
				}
			} else if request.Status == "reject" {
				// อัปเดตสถานะ Shipment
				shipment.Status = "Rejected"
				if err := tx.Save(&shipment).Error; err != nil {
					return fmt.Errorf("failed to update shipment status: %v", err)
				}
			}
			return nil
		}); err != nil {
			log.Printf("Error syncing request %s with warehouse: %v\n", request.RequestID, err)
			continue
		}

		log.Printf("Request %s synced successfully with Warehouse\n", request.RequestID)
	}

	return nil
}

func LookShipments(db *gorm.DB, c *fiber.Ctx) error {
	var shipments []Models.Shipment
	if err := db.Find(&shipments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch shipments",
		})
	}

	// เพิ่ม Log เพื่อดูข้อมูลที่ส่งกลับ
	log.Println("Shipments fetched:", shipments)

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
