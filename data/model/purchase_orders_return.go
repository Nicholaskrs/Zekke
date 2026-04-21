package model

type PurchaseOrderReturn struct {
	ID                    uint       `gorm:"primaryKey;column:id"`
	Qty                   int        `gorm:"column:qty;type:int"`
	Remarks               string     `gorm:"column:remarks;type:text"`
	PurchaseOrderID       uint       `gorm:"column:purchase_order_id;type:int"`
	PurchaseOrderDetailID uint       `gorm:"column:purchase_order_detail_id;type:int"`
	CreatedByID           uint       `gorm:"column:created_by_id;type:int"`
	Timestamp             *Timestamp `gorm:"embedded"`
}

func (PurchaseOrderReturn) TableName() string {
	return "purchase_order_returns"
}
