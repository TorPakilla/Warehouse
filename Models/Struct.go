package Models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Employees model
type Employees struct {
	EmployeesID string    `gorm:"type:uuid;primaryKey" json:"employees_id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	BranchID    string    `gorm:"column:branch_id;foreignKey:BranchID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Salary      float64   `json:"salary"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Employees) TableName() string {
	return "Employees"
}

func (s *Employees) BeforeCreate(tx *gorm.DB) (err error) {
	s.EmployeesID = uuid.New().String()
	return
}

// Branch model
type Branches struct {
	BranchID  string    `gorm:"type:uuid;primaryKey" json:"branch_id"`
	BName     string    `json:"b_name"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
}

func (Branches) TableName() string {
	return "Branches"
}

func (b *Branches) BeforeCreate(tx *gorm.DB) (err error) {
	b.BranchID = uuid.New().String()
	return
}

// Product model
type Product struct {
	ProductID   string    `gorm:"type:uuid;primaryKey" json:"product_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	Image       []byte    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Product) TableName() string {
	return "Product"
}

func (s *Product) BeforeCreate(tx *gorm.DB) (err error) {
	s.ProductID = uuid.New().String()
	return
}

// ProductUnit model
type ProductUnit struct {
	ProductUnitID   string    `gorm:"primaryKey;column:product_unit_id" json:"product_unit_id"`
	ProductID       string    `gorm:"column:product_id" json:"product_id"`
	Type            string    `gorm:"column:type" json:"type"`
	InitialQuantity int       `gorm:"column:initial_quantity" json:"initial_quantity"`
	ConversRate     int       `gorm:"column:convers_rate" json:"convers_rate"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ProductUnit) TableName() string {
	return "ProductUnit"
}

func (p *ProductUnit) BeforeCreate(tx *gorm.DB) (err error) {
	p.ProductUnitID = uuid.New().String()
	return
}

// Inventory model
type Inventory struct {
	InventoryID string    `gorm:"primaryKey;column:inventory_id" json:"inventory_id"`
	ProductID   string    `gorm:"column:product_id" json:"product_id"`
	BranchID    string    `gorm:"column:branch_id" json:"branch_id"`
	Quantity    int       `gorm:"column:quantity" json:"quantity"`
	Price       float64   `gorm:"column:price" json:"price"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Inventory) TableName() string {
	return "Inventory"
}

func (s *Inventory) BeforeCreate(tx *gorm.DB) (err error) {
	s.InventoryID = uuid.New().String()
	return
}

// Supplier model
type Supplier struct {
	SupplierID  string  `gorm:"type:uuid;primaryKey" json:"supplier_id"`
	Name        string  `json:"name"`
	ProductID   string  `gorm:"type:uuid" json:"product_id"`
	PricePallet float64 `json:"price_pallet"`
}

func (Supplier) TableName() string {
	return "Supplier"
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) (err error) {
	s.SupplierID = uuid.New().String()
	return
}

type Order struct {
	OrderID     string     `gorm:"type:uuid;primaryKey" json:"order_id"`
	OrderNumber string     `json:"order_number"`
	Status      string     `json:"status"`
	SupplierID  uuid.UUID  `gorm:"type:uuid" json:"supplier_id"`
	EmployeesID *uuid.UUID `gorm:"type:uuid" json:"employees_id"`
	TotalAmount float64    `json:"total_amount"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	Employees  *Employees  `gorm:"foreignKey:EmployeesID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"employees"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`
}

func (Order) TableName() string {
	return `"Order"`
}

// BeforeCreate sets a new UUID before creating the record
func (s *Order) BeforeCreate(tx *gorm.DB) (err error) {
	s.OrderID = uuid.New().String()
	return
}

type OrderItem struct {
	OrderItemID string    `gorm:"type:uuid;primaryKey" json:"order_item_id"`
	OrderID     string    `gorm:"type:uuid;not null" json:"order_id"`
	ProductID   string    `gorm:"type:uuid;not null" json:"product_id"`
	Quantity    int       `json:"quantity"`
	ConversRate float64   `json:"convers_rate"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (OrderItem) TableName() string {
	return "OrderItem"
}

// BeforeCreate sets a new UUID before creating the record
func (s *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	s.OrderItemID = uuid.New().String()
	return
}

// Shipment model
type Shipment struct {
	ShipmentID     string    `gorm:"type:uuid;primaryKey" json:"shipment_id"`
	ShipmentNumber string    `json:"shipment_number"`
	FromBranchID   string    `gorm:"type:uuid" json:"from_branch_id"`
	ToBranchID     string    `gorm:"type:uuid" json:"to_branch_id"`
	ShipmentDate   time.Time `json:"shipment_date"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Shipment) TableName() string {
	return "Shipment"
}

func (s *Shipment) BeforeCreate(tx *gorm.DB) (err error) {
	s.ShipmentID = uuid.New().String()
	return
}

// ShipmentItem model
type ShipmentItem struct {
	ShipmentListID       string    `gorm:"type:uuid;primaryKey" json:"shipment_list_id"`
	ShipmentID           string    `gorm:"type:uuid" json:"shipment_id"`
	ProductUnitID        string    `gorm:"type:uuid" json:"product_unit_id"`
	WarehouseInventoryID string    `gorm:"type:uuid" json:"warehouse_inventory_id"`
	PosInventoryID       string    `gorm:"type:uuid" json:"pos_inventory_id"`
	Status               string    `json:"status"`
	Quantity             int       `json:"quantity"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (ShipmentItem) TableName() string {
	return "ShipmentItem"
}

func (s *ShipmentItem) BeforeCreate(tx *gorm.DB) (err error) {
	s.ShipmentListID = uuid.New().String()
	return
}
