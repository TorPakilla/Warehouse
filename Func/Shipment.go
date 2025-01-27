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

// AddShipment handles adding a new shipment along with requests
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
	log.Printf("Generated ShipmentID: %s", shipmentID.String())

	if err := db.Transaction(func(tx *gorm.DB) error {
		shipment := Models.Shipment{
			ShipmentID:     shipmentID.String(),
			ShipmentNumber: GenerateULID(),
			FromBranchID:   req.FromBranchID,
			ToBranchID:     req.ToBranchID,
			Status:         "Pending",
			ShipmentDate:   time.Now(),
		}

		log.Printf("Before saving shipment: %+v", shipment)

		if err := tx.Save(&shipment).Error; err != nil {
			log.Println("Error creating shipment:", err)
			return err
		}

		log.Printf("After saving shipment: %+v", shipment)

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
				log.Println("Error finding PosInventory:", err)
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

			log.Printf("Creating shipment item: %+v", shipmentItem)

			if err := tx.Create(&shipmentItem).Error; err != nil {
				log.Println("Error creating shipment item:", err)
				return err
			}

			request := Request{
				RequestID:    shipmentID,
				FromBranchID: req.FromBranchID,
				ToBranchID:   req.ToBranchID,
				ProductID:    posInventory.ProductID.String(),
				Quantity:     int(quantity),
				Status:       "pending",
				CreatedAt:    time.Now(),
			}

			log.Printf("Creating request: %+v", request)

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
func UpdateShipment(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id") // รับ Shipment ID
	var shipment Models.Shipment

	// Fetch Shipment
	if err := db.Where("shipment_id = ?", id).First(&shipment).Error; err != nil {
		log.Println("Error finding shipment:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shipment not found"})
	}

	type ShipmentRequest struct {
		Status string `json:"status"` // รับสถานะใหม่
	}

	var req ShipmentRequest
	if err := c.BodyParser(&req); err != nil {
		log.Println("Error parsing request:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// Validate Status
	allowedStatuses := map[string]bool{"Pending": true, "Approved": true, "Rejected": true}
	if !allowedStatuses[req.Status] {
		log.Println("Invalid status provided:", req.Status)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	// เมื่อสถานะถูกเปลี่ยนเป็น Approved
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
				// Debug การอัปเดต Inventory
				log.Printf("Processing item: %+v\n", item)

				// Fetch Inventory ก่อนการอัปเดต
				var inventory Inventory
				if err := tx.Where("inventory_id = ?", item.WarehouseInventoryID).First(&inventory).Error; err != nil {
					log.Printf("Error finding inventory for ID %s: %v\n", item.WarehouseInventoryID, err)
					return fmt.Errorf("failed to find inventory for item: %s", item.WarehouseInventoryID)
				}

				log.Printf("Before update - Inventory ID: %s, Current Quantity: %d, Required Quantity: %d",
					item.WarehouseInventoryID, inventory.Quantity, item.Quantity)

				// ลดสินค้าใน Inventory
				if err := tx.Model(&Inventory{}).
					Where("inventory_id = ?", item.WarehouseInventoryID).
					Where("quantity >= ?", item.Quantity).
					Update("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
					log.Printf("Failed to decrease Warehouse Inventory for ID %s: %v", item.WarehouseInventoryID, err)
					return fmt.Errorf("failed to update Warehouse Inventory for item: %s", item.WarehouseInventoryID)
				}

				// ตรวจสอบ Inventory หลังจากการอัปเดต
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

func AutoUpdateShipments(db *gorm.DB) error {
	var shipments []Models.Shipment

	log.Printf("Fetching shipments with status 'Approved'...")
	// ดึงข้อมูล Shipments ที่สถานะ Approved
	if err := db.Where("status = ?", "Approved").Find(&shipments).Error; err != nil {
		log.Println("Error fetching approved shipments for auto-update:", err)
		return err
	}
	log.Printf("Found %d shipments with status 'Approved'", len(shipments))

	for _, shipment := range shipments {
		log.Printf("Processing shipment ID: %s", shipment.ShipmentID)

		// ดึง ShipmentItems ที่เกี่ยวข้อง
		var shipmentItems []Models.ShipmentItem
		if err := db.Where("shipment_id = ?", shipment.ShipmentID).Find(&shipmentItems).Error; err != nil {
			log.Printf("Error fetching shipment items for shipment ID %s: %v", shipment.ShipmentID, err)
			continue
		}
		log.Printf("Found %d shipment items for shipment ID: %s", len(shipmentItems), shipment.ShipmentID)

		// อัปเดต Inventory
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, item := range shipmentItems {
				log.Printf("Processing shipment item: %+v", item)

				// ลดจำนวนสินค้าใน Warehouse Inventory
				if err := tx.Debug().Model(&Inventory{}).
					Where("inventory_id = ?", item.WarehouseInventoryID).
					Where("quantity >= ?", item.Quantity). // ตรวจสอบว่ามีสินค้าเพียงพอ
					Update("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
					log.Printf("Failed to update inventory for WarehouseInventoryID %s: %v", item.WarehouseInventoryID, err)
					return fmt.Errorf("not enough inventory for WarehouseInventoryID %s", item.WarehouseInventoryID)
				}

				log.Printf("Inventory updated for WarehouseInventoryID: %s, Quantity Decreased: %d", item.WarehouseInventoryID, item.Quantity)
			}

			// อัปเดตสถานะ Shipment เป็น "Completed"
			shipment.Status = "Completed"
			shipment.UpdatedAt = time.Now()
			if err := tx.Save(&shipment).Error; err != nil {
				log.Printf("Failed to update shipment ID %s: %v", shipment.ShipmentID, err)
				return err
			}

			log.Printf("Shipment ID %s updated to 'Completed'", shipment.ShipmentID)
			return nil
		}); err != nil {
			log.Printf("Error processing shipment ID %s: %v", shipment.ShipmentID, err)
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
		log.Println("Error fetching requests from POS:", err)
		return err
	}

	for _, request := range requests {
		// ข้าม RequestID ที่เคยหาไม่เจอในรอบก่อนหน้า
		if notFoundRequests[request.RequestID] {
			continue // ไม่แสดง Log ใด ๆ สำหรับรายการที่ไม่พบ
		}

		var shipment Models.Shipment

		// ค้นหา Shipment ที่ตรงกับ RequestID
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

				// อัปเดต Inventory (ลดจากต้นทาง เพิ่มในปลายทาง)
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

	// ซิงค์สถานะระหว่าง POS และ Warehouse ทุก 10 วินาที
	scheduler.Every(10).Seconds().Do(func() {
		log.Println("Starting request sync task...")
		if err := SyncRequestStatusWithWarehouse(db, posDB); err != nil {
			log.Printf("Error syncing requests: %v\n", err)
		} else {
			log.Println("Request sync task completed successfully.")
		}
	})

	// ลดจำนวนสินค้าใน Inventory ทุก 10 นาที สำหรับ Shipments ที่ Approved
	scheduler.Every(12).Seconds().Do(func() {
		log.Println("Starting shipment auto-update task...")
		if err := AutoUpdateShipments(db); err != nil {
			log.Printf("Error updating shipments: %v\n", err)
		} else {
			log.Println("Shipment auto-update task completed successfully.")
		}
	})

	// เริ่ม Scheduler
	scheduler.StartAsync()
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
