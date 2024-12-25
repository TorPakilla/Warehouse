package Models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Emloyees struct {
	EmloyeeID string    `gorm:"type:uuid;primaryKey" json:"emloyeeid"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	BrancheID string    `gorm:"type:uuid;foreignKey:BrancheID" json:"brancheid"`
	Salary    float64   `json:"salary"`
	CreatedAt uuid.Time `json:"createdat"`
}

func (Emloyees) TableName() string {
	return "Emloyees"
}

func (s *Emloyees) BeforeCreate(tx *gorm.DB) (err error) {
	s.EmloyeeID = uuid.New().String()
	return
}

type Branches struct {
	BrancheID string     `gorm:"type:uuid;primaryKey" json:"brancheid"`
	BName     string     `json:"bname"`
	Location  string     `json:"location"`
	CreatedAt uuid.Time  `json:"createdat"`
	Emloyees  []Emloyees `gorm:"foreignKey:BrancheID" json:"emloyees"`
}

func (Branches) TableName() string {
	return "Branches"
}

func (s *Branches) BeforeCreate(tx *gorm.DB) (err error) {
	s.BrancheID = uuid.New().String()
	return
}

type Product struct {
	ProductID   string    `gorm:"type:uuid;primaryKey" json:"productid"`
	ProductName string    `json:"productname"`
	Description string    `json:"description"`
	CreatedAt   uuid.Time `json:"createdat"`
}

func (Product) TableName() string {
	return "Product"
}

func (s *Product) BeforeCreate(tx *gorm.DB) (err error) {
	s.ProductID = uuid.New().String()
	return
}

type ProductUnit struct {
	ProductUnitID string    `gorm:"type:uuid;primaryKey" json:"productunitid"`
	ProductID     []Product `gorm:"foreignKey:ProductID" json:"productid"`
	Type          string    `json:"type"`
	ConversRate   int       `json:"conversrate"`
}

func (ProductUnit) TableName() string {
	return "ProductUnit"
}

func (s *ProductUnit) BeforeCreate(tx *gorm.DB) (err error) {
	s.ProductUnitID = uuid.New().String()
	return
}

type Inventory struct {
	InventoryID   string        `gorm:"type:uuid;primaryKey" json:"inventoryid"`
	ProductUnitID []ProductUnit `gorm:"foreignKey:ProductUnitID" json:"productunitid"`
	BrancheID     string        `gorm:"type:uuid;foreignKey:BrancheID" json:"brancheid"`
	Quantity      int           `json:"quantity"`
	Price         float64       `json:"price"`
	CreatedAt     uuid.Time     `json:"createdat"`
}

func (Inventory) TableName() string {
	return "Inventory"
}

func (s *Inventory) BeforeCreate(tx *gorm.DB) (err error) {
	s.InventoryID = uuid.New().String()
	return
}

type Supplier struct {
	SupplierID  string    `gorm:"type:uuid;primaryKey" json:"supplierid"`
	Name        string    `json:"name"`
	ProductID   []Product `gorm:"foreignKey:ProductID" json:"productid"`
	PricePallet int       `json:"pricepallet"`
}

func (Supplier) TableName() string {
	return "Supplier"
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) (err error) {
	s.SupplierID = uuid.New().String()
	return
}

type Order struct {
	OrderID     string    `gorm:"type:uuid;primaryKey" json:"orderid"`
	OrderNumber string    `json:"ordernumber"`
	EmplyeeID   string    `gorm:"type:uuid;foreignKey:EmloyeeID" json:"emplyeeid"`
	SupplierID  string    `gorm:"type:uuid;foreignKey:SupplierID" json:"supplierid"`
	ReceiptDate uuid.Time `json:"receiptdate"`
	TotalAmount float64   `json:"totalamount"`
	Status      string    `json:"status"`
	CreatedAt   uuid.Time `json:"createdat"`
	UpdateAt    uuid.Time `json:"updateat"`
}

func (Order) TableName() string {
	return "Order"
}

func (s *Order) BeforeCreate(tx *gorm.DB) (err error) {
	s.OrderID = uuid.New().String()
	return
}

type OrderItem struct {
	OrderItemID   string    `gorm:"type:uuid;primaryKey" json:"orderitemid"`
	OrderID       string    `gorm:"type:uuid;foreignKey:OrderID" json:"orderid"`
	ProductUnitID string    `gorm:"type:uuid;foreignKey:ProductUnitID" json:"productunitid"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unitprice"`
	CreatedAt     uuid.Time `json:"createdat"`
	UpdateAt      uuid.Time `json:"updateat"`
}

func (OrderItem) TableName() string {
	return "OrderItem"
}

func (s *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	s.OrderItemID = uuid.New().String()
	return
}

type Shipment struct {
	ShipmentID     string    `gorm:"type:uuid;primaryKey" json:"shipmentid"`
	ShipmentNumber string    `json:"shipmentnumber"`
	FromBranchID   string    `gorm:"type:uuid;foreignKey:BrancheID" json:"frombranchid"`
	ShipmentDate   uuid.Time `json:"shipmentdate"`
	Status         string    `json:"status"`
	CreatedAt      uuid.Time `json:"createdat"`
	UpdateAt       uuid.Time `json:"updateat"`
}

func (Shipment) TableName() string {
	return "Shipment"
}

func (s *Shipment) BeforeCreate(tx *gorm.DB) (err error) {
	s.ShipmentID = uuid.New().String()
	return
}

type ShipmentItem struct {
	ShipmentListID string    `gorm:"type:uuid;primaryKey" json:"shipmentlistid"`
	ShipmentID     string    `gorm:"type:uuid;foreignKey:ShipmentID" json:"shipmentid"`
	ProductUnitID  string    `gorm:"type:uuid;foreignKey:ProductUnitID" json:"productunitid"`
	Quantity       int       `json:"quantity"`
	CreatedAt      uuid.Time `json:"createdat"`
	UpdateAt       uuid.Time `json:"updateat"`
}

func (ShipmentItem) TableName() string {
	return "ShipmentItem"
}

func (s *ShipmentItem) BeforeCreate(tx *gorm.DB) (err error) {
	s.ShipmentListID = uuid.New().String()
	return
}
