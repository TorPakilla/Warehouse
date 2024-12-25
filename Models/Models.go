package Models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Users struct {
	UserID   string   `gorm:"type:uuid;primaryKey" json:"userid"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Name     string   `json:"name"`
	Role     string   `json:"role"`
	Branche  string   `gorm:"type:uuid;not null" json:"branche"`
	Branches Branches `gorm:"foreignKey:Branche" json:"branches"`
}

func (Users) TableName() string {
	return "Users"
}

func (s *Users) BeforeCreate(tx *gorm.DB) (err error) {
	s.UserID = uuid.New().String()
	return
}

type Branches struct {
	BrancheID string  `gorm:"type:uuid;primaryKey" json:"id"`
	BName     string  `json:"bname"`
	Location  string  `json:"location"`
	Branches  []Users `gorm:"foreignKey:BrancheID" json:"branche"`
}

func (Branches) TableName() string {
	return "Branches"
}

func (s *Branches) BeforeCreate(tx *gorm.DB) (err error) {
	s.BrancheID = uuid.New().String()
	return
}

type Pallets struct {
	PalletID string  `gorm:"type:uuid;primaryKey" json:"palletid"`
	BName    string  `json:"bname"`
	Location string  `json:"location"`
	Branches []Users `gorm:"foreignKey:BrancheID" json:"branche"`
}

func (Branches) TableName() string {
	return "Branches"
}

func (s *Branches) BeforeCreate(tx *gorm.DB) (err error) {
	s.BrancheID = uuid.New().String()
	return
}

type Supply struct {
	ID          string   `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	UnitBox     int      `json:"unitbox"`
	SupplierID  string   `gorm:"type:uuid;not null" json:"supplier_id"`
	Supplier    Supplier `gorm:"foreignKey:SupplierID" json:"supplier"`
}

func (Supply) TableName() string {
	return "Supply"
}

func (s *Supply) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}

type Supplier struct {
	ID          string   `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string   `gorm:"uniqueIndex;not null" json:"name"`
	Description string   `json:"description"`
	Supplies    []Supply `gorm:"foreignKey:SupplierID" json:"supplies"`
}

func (Supplier) TableName() string {
	return "Supplier"
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}
