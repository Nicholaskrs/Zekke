package model

import "github.com/shopspring/decimal"

type DistributorProduct struct {
	ID            uint            `gorm:"primaryKey;column:id"`
	ProductID     uint            `gorm:"column:product_id;type:int"`
	DistributorID uint            `gorm:"column:distributor_id;type:int"`
	Qty           int             `gorm:"column:qty;type:int"`
	Price         decimal.Decimal `gorm:"column:price;type:decimal(20,4)"`
}

func (DistributorProduct) TableName() string {
	return "distributor_products"
}
