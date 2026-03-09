package models

import "time"

type SalesOrderLineTax struct {
	OrdrltxID uint64    `gorm:"column:ordrltx_id;primaryKey"`
	OrdrlID   uint64    `gorm:"column:ordrl_id;not null"`
	TaxType   string    `gorm:"column:tax_type;not null"`
	TaxCode   string    `gorm:"column:tax_code;not null"`
	TaxAmount float64   `gorm:"column:tax_amount;type:numeric(15,3);not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (SalesOrderLineTax) TableName() string {
	return "sales_order_line_taxes"
}
