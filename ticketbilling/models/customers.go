package models

import "time"

type Customer struct {
	CustID     uint64    `gorm:"column:cust_id;primaryKey"`
	TaxID      string    `gorm:"column:tax_id;not null"`
	LegalName  string    `gorm:"column:legal_name;not null"`
	TaxRegime  *string   `gorm:"column:tax_regime"`
	PostalCode *string   `gorm:"column:postal_code"`
	Email      *string   `gorm:"column:email"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Customer) TableName() string {
	return "customers"
}
