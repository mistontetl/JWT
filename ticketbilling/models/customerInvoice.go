package models

import "time"

type CustomerInvoice struct {
	/*
		InvcID       uint64     `gorm:"column:invc_id;primaryKey"`
		OrdrID       uint64     `gorm:"column:ordr_id;not null"`
		InvsID       *uint64    `gorm:"column:invs_id"`
		Folio        *string    `gorm:"column:folio"`
		InvoicedAt   *time.Time `gorm:"column:invoiced_at"`
		CFDIUUID     *string    `gorm:"column:cfdi_uuid"`
		DigitalSeal  *string    `gorm:"column:digital_seal;type:text"`
		XMLCFDI      *string    `gorm:"column:xml_cfdi;type:text"`
		PDFCFDI      []byte     `gorm:"column:pdf_cfdi"`
		ErrorCode    *string    `gorm:"column:error_code"`
		ErrorMessage *string    `gorm:"column:error_message;type:text"`
		AttemptedAt  time.Time  `gorm:"column:attempted_at"`
		IsInvoiced   bool       `gorm:"column:is_invoiced"`
		CreatedAt    time.Time  `gorm:"column:created_at"`
		SourceSaleID *string    `gorm:"column:source_sale_id"`
		SourceSystem *string    `gorm:"column:source_system"`*/
	InvcID         uint64     `gorm:"column:invc_id;primaryKey"`
	Folio          *string    `gorm:"column:folio"`
	InvoicedAt     *time.Time `gorm:"column:invoiced_at"`
	CFDIUUID       *string    `gorm:"column:cfdi_uuid"`
	DigitalSeal    *string    `gorm:"column:digital_seal;type:text"`
	XMLCFDI        *string    `gorm:"column:xml_cfdi;type:text"`
	PDFCFDI        []byte     `gorm:"column:pdf_cfdi"`
	ErrorCode      *string    `gorm:"column:error_code"`
	ErrorMessage   *string    `gorm:"column:error_message;type:text"`
	AttemptedAt    time.Time  `gorm:"column:attempted_at"`
	IsInvoiced     bool       `gorm:"column:is_invoiced"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	SourceTicketID string     `gorm:"column:source_ticket_id;not null"`
	SourceSystem   string     `gorm:"column:source_system;not null"`
}

func (CustomerInvoice) TableName() string {
	return "customer_invoice"
}
