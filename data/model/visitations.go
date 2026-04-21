package model

import "template-go/data/enum"

type Visitation struct {
	ID                      uint                       `gorm:"primaryKey;column:id"`
	StoreID                 uint                       `gorm:"column:store_id;type:int"`
	Latitude                float64                    `gorm:"column:latitude;type:decimal(9,6)"`
	Longitude               float64                    `gorm:"column:longitude;type:decimal(9,6)"`
	GoodsCondition          string                     `gorm:"column:goods_condition;type:string"`
	GoodsNotes              string                     `gorm:"column:defective_goods_count;type:string"`
	VariantCount            int                        `gorm:"column:variant_count;type:int"`
	FreezerState            enum.FreezerState          `gorm:"column:freezer_state;type:ENUM('Clean', 'Not Clean', 'Dirty')"`
	FreezerThermometerState enum.FreezerThermometer    `gorm:"column:freezer_thermometer_state;type:ENUM('Good', 'Bad')"`
	FreezerGapState         enum.FreezerState          `gorm:"column:freezer_gap_state;type:ENUM('Clean', 'Not Clean', 'Dirty')"`
	FreezerGlassState       enum.FreezerState          `gorm:"column:freezer_glass_state;type:ENUM('Clean', 'Not Clean', 'Dirty')"`
	FreezerPositioning      enum.FreezerPositioning    `gorm:"column:freezer_positioning;type:ENUM('Optimal', 'Acceptable', 'Need Adjustment')"`
	FreezerCapacityUpper    enum.FreezerCapacityStatus `gorm:"column:freezer_capacity_upper;type:ENUM('0-20', '20-40', '40-60', '60-80', '80-100')"`
	FreezerCapacityLower    enum.FreezerCapacityStatus `gorm:"column:freezer_capacity_lower;type:ENUM('0-20', '20-40', '40-60', '60-80', '80-100')"`
	IsPriceBoardDisplayed   bool                       `gorm:"column:is_price_board_displayed"`
	IsPriceStickerDisplayed bool                       `gorm:"column:is_price_sticker_displayed"`
	IsBannerDisplayed       bool                       `gorm:"column:is_banner_displayed"`
	IsPosterDisplayed       bool                       `gorm:"column:is_poster_displayed"`
	IsFlagDisplayed         bool                       `gorm:"column:is_flag_displayed"`
	CreatedByID             uint                       `gorm:"column:created_by_id;type:int"`
	Timestamp               *Timestamp                 `gorm:"embedded"`
}

func (Visitation) TableName() string {
	return "visitations"
}
