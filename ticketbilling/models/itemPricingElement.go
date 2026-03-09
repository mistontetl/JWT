package models

import (
	"time"
)

type ItemPricingElement struct {
	ItpeID                   uint64    `gorm:"column:itpe_id;primaryKey"`
	OrdrlID                  uint64    `gorm:"column:ordrl_id;not null"`
	ConditionType            string    `gorm:"column:condition_type;not null;default:PMP0"`
	ConditionRateAmount      float64   `gorm:"column:condition_rate_amount;type:numeric(15,3);not null"`
	ConditionCurrency        string    `gorm:"column:condition_currency;not null;default:MXN"`
	ConditionQuantity        float64   `gorm:"column:condition_quantity;type:numeric(15,3);not null"`
	ConditionQuantityISOUnit string    `gorm:"column:condition_quantity_iso_unit;not null;default:EA"`
	CreatedAt                time.Time `gorm:"column:created_at"`

	SalesOrderLine SalesOrderLine `gorm:"foreignKey:OrdrlID"`
}

func (ItemPricingElement) TableName() string {
	return "item_pricing_element"
}
