package Models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderSupply struct {
	ID          string   `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	UnitBox     int      `json:"unitbox"`
	SupplierID  string   `gorm:"type:uuid;not null" json:"supplier_id"`
	Supplier    Supplier `gorm:"foreignKey:SupplierID" json:"supplier"`
}

func (OrderSupply) TableName() string {
	return "OrderSupply"
}

func (s *OrderSupply) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}

type Supplier struct {
	ID          string        `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string        `gorm:"uniqueIndex;not null" json:"name"`
	Description string        `json:"description"`
	Supplies    []OrderSupply `gorm:"foreignKey:SupplierID" json:"supplies"`
}

func (Supplier) TableName() string {
	return "Supplier"
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}
