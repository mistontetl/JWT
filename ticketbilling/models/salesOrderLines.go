package models

import "time"

type SalesOrderLine struct {
	OrdrlID              uint64    `gorm:"column:ordrl_id;primaryKey"`
	OrdrID               uint64    `gorm:"column:ordr_id;not null"`
	LineNumber           int       `gorm:"column:line_number;not null"`
	Description          string    `gorm:"column:description;type:text;not null"`
	Quantity             float64   `gorm:"column:quantity;type:numeric(15,3);not null"`
	UnitPrice            float64   `gorm:"column:unit_price;type:numeric(15,3);not null"`
	LineAmount           float64   `gorm:"column:line_amount;type:numeric(15,3);not null"`
	DiscountAmount       float64   `gorm:"column:discount_amount;type:numeric(15,3)"`
	ProductSATCode       string    `gorm:"column:product_sat_code;not null"`
	UnitSATCode          string    `gorm:"column:unit_sat_code;not null"`
	SKU                  *string   `gorm:"column:sku"`
	ProductSAPMaterial   *string   `gorm:"column:product_sap_material"`
	QuantityUOM          *string   `gorm:"column:quantity_uom"`
	ProfitCenter         *string   `gorm:"column:profit_center"`
	ConditionType        *string   `gorm:"column:condition_type"`
	ConditionQuantityUOM *string   `gorm:"column:condition_quantity_uom"`
	ConditionRateAmount  float64   `gorm:"column:condition_rate_amount;type:numeric(15,3)"`
	ConditionCurrency    *string   `gorm:"column:condition_currency"`
	ConditionQuantity    float64   `gorm:"column:condition_quantity;type:numeric(15,3)"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (SalesOrderLine) TableName() string {
	return "sales_order_lines"
}
