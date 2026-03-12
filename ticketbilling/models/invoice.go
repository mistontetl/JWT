package models

import "time"

type Invoice struct {
	InsID            uint       `gorm:"column:ins_id;primaryKey;autoIncrement"`
	CadenaCFDI       string     `gorm:"column:cadena_cfdi"`
	CertificadoSAT   string     `gorm:"column:certificado_sat"`
	DateCreate       time.Time  `gorm:"column:date_create"`
	FechaCancelacion *time.Time `gorm:"column:fecha_cancelacion"`
	FechaCFDI        time.Time  `gorm:"column:fecha_cfdi"`
	FolioFactura     string     `gorm:"column:folio_factura"`
	FolioFiscal      string     `gorm:"column:folio_fiscal"`
	PaymentMethod    string     `gorm:"column:payment_method"`
	PDFFile          string     `gorm:"column:pdf_file"`
	PDFXML           string     `gorm:"column:pdf_xml"`
	SelloCFDI        string     `gorm:"column:sello_cfdi"`
	SelloSAT         string     `gorm:"column:sello_sat"`
	Total            float64    `gorm:"column:total"`
	VersionTimbre    string     `gorm:"column:version_timbre"`
}

func (Invoice) TableName() string {
	return "invoices"
}
