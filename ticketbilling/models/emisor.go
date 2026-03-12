package models

import "time"

type Emisor struct {
	IssrID     uint64    `gorm:"column:issr_id;primaryKey"`
	TaxID      string    `gorm:"column:tax_id;not null"`
	LegalName  string    `gorm:"column:legal_name;not null"`
	TaxRegime  string    `gorm:"column:tax_regime;not null"`
	PostalCode string    `gorm:"column:postal_code;not null"`
	Active     bool      `gorm:"column:active"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Emisor) TableName() string {
	return "emisor"
}
