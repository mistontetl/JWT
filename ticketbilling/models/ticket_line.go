package models

import "time"

type TicketLine struct {
	TklID                int64      `gorm:"column:tkl_id;primaryKey;not null"`
	TkID                 int64      `gorm:"column:tk_id;not null"`
	LineNumber           int        `gorm:"column:line_number;not null"`
	Description          string     `gorm:"column:description;type:text;not null"`
	Quantity             float64    `gorm:"column:quantity;type:numeric(15,3);not null"`
	UnitPrice            float64    `gorm:"column:unit_price;type:numeric(15,3);not null"`
	LineAmount           float64    `gorm:"column:line_amount;type:numeric(15,3);not null"`
	DiscountAmount       *float64   `gorm:"column:discount_amount;type:numeric(15,3)"`
	ProductSATCode       string     `gorm:"column:product_sat_code;not null"`
	UnitSATCode          string     `gorm:"column:unit_sat_code;not null"`
	SKU                  *string    `gorm:"column:sku"`
	ProductSAPMaterial   *string    `gorm:"column:product_sap_material"`
	QuantityUOM          *string    `gorm:"column:quantity_uom"`
	ProfitCenter         *string    `gorm:"column:profit_center"`
	ConditionType        *string    `gorm:"column:condition_type"`
	ConditionQuantityUOM *string    `gorm:"column:condition_quantity_uom"`
	CreatedAt            *time.Time `gorm:"column:created_at"`
	ConditionRateAmount  *float64   `gorm:"column:condition_rate_amount;type:numeric(15,3)"`
	ConditionCurrency    *string    `gorm:"column:condition_currency"`
	ConditionQuantity    *float64   `gorm:"column:condition_quantity;type:numeric(15,3)"`

	TaxRate         float64 `gorm:"-"`
	TaxRateTypeCode string  `gorm:"-"`
}

func (TicketLine) TableName() string {
	return "ticket_lines"
}

/*
type TicketLine struct {
	TklID               uint       `gorm:"column:tkl_id;primaryKey;autoIncrement"`
	Amount              float64    `gorm:"column:amount"`
	Base                float64    `gorm:"column:base"`
	Cantidad            float64    `gorm:"column:cantidad"`
	ClaveProdServ       string     `gorm:"column:clave_prod_serv"`
	ClaveUnidad         string     `gorm:"column:clave_unidad"`
	DateCreate          *time.Time `gorm:"column:date_create"`
	Descripcion         string     `gorm:"column:descripcion"`
	Descuento           string     `gorm:"column:descuento"`
	NoIdentificacion    string     `gorm:"column:no_identificacion"`
	PorcentajeDescuento string     `gorm:"column:porcentaje_descuento"`
	TaxRate             float64    `gorm:"column:taxrate"`
	TaxRateTypeCode     string     `gorm:"column:tax_rate_type_code"`
	ValorUnitario       float64    `gorm:"column:valor_unitario"`

	//FK
	TkID uint `gorm:"column:tk_id"`
}

func (TicketLine) TableName() string {
	return "ticket_lines"
}
*/
