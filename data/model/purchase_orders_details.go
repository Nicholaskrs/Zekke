package model

import "github.com/shopspring/decimal"

type PurchaseOrderDetail struct {
	ID              uint            `gorm:"primaryKey;column:id"`
	PurchaseOrderID uint            `gorm:"column:purchase_order_id;type:int"`
	ProductID       uint            `gorm:"column:product_id;type:int"`
	Qty             int             `gorm:"column:qty;type:int"`
	Price           decimal.Decimal `gorm:"column:price;type:decimal(20,4)"`
	Timestamp       *Timestamp      `gorm:"embedded"`
}

func (PurchaseOrderDetail) TableName() string {
	return "purchase_order_details"
}
