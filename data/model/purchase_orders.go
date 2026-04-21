package model

import (
	"github.com/shopspring/decimal"
	"template-go/data/enum"
)

type PurchaseOrder struct {
	ID                  uint             `gorm:"primaryKey;column:id"`
	StoreID             uint             `gorm:"column:store_id;type:int"`
	VisitationID        *uint            `gorm:"column:visitation_id;type:int;null"`
	Status              enum.OrderStatus `gorm:"column:status;type:varchar(255);type:ENUM('Created', 'Processed', 'Rejected', 'Delivered')"`
	RejectionRemark     *string          `gorm:"column:rejection_remark;type:varchar(255);null"`
	DeliveryOrderImgUrl *string          `gorm:"column:delivery_order_img_url;type:varchar(255);null"`
	Date                string           `gorm:"column:date;type:date;null"`
	TotalSales          decimal.Decimal  `gorm:"column:total_sales;type:decimal(20,8);null"`
	TotalQty            int              `gorm:"column:total_qty;type:int"`
	CreatedByID         uint             `gorm:"column:created_by_id;type:int"`
	Timestamp           *Timestamp       `gorm:"embedded"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}
