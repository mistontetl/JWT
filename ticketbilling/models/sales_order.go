package models

import "time"

type SalesOrder struct {
	OrdrID                    int64      `gorm:"column:tk_id"`
	SourceSaleUUID            string     `gorm:"column:source_ticket_uuid"`
	SourceSaleID              string     `gorm:"column:source_ticket_id"`
	SourceSystem              string     `gorm:"column:source_system"`
	IssrID                    int64      `gorm:"column:issr_id"`
	CustID                    int64      `gorm:"column:cust_id"`
	IssueDatetime             time.Time  `gorm:"column:issue_datetime"`
	Currency                  string     `gorm:"column:currency"`
	PaymentForm               string     `gorm:"column:payment_form"`
	Subtotal                  float64    `gorm:"column:subtotal"`
	TotalDiscounts            float64    `gorm:"column:total_discounts"`
	TotalTaxTransferred       float64    `gorm:"column:total_tax_transferred"`
	TotalTaxWithheld          float64    `gorm:"column:total_tax_withheld"`
	TotalAmount               float64    `gorm:"column:total_amount"`
	SalesOrderType            string     `gorm:"column:sales_order_type"`
	SoldToParty               string     `gorm:"column:sold_to_party"`
	SalesOrganization         string     `gorm:"column:sales_organization"`
	DistributionChannel       string     `gorm:"column:distribution_channel"`
	OrganizationDivision      string     `gorm:"column:organization_division"`
	PurchaseOrderByCustomer   string     `gorm:"column:purchase_order_by_customer"`
	CustomerPurchaseOrderDate *time.Time `gorm:"column:customer_purchase_order_date"`
	RequestedDeliveryDate     *time.Time `gorm:"column:requested_delivery_date"`
	PaymentMethodSAP          string     `gorm:"column:payment_method_sap"`
	YY1EventoSDH              string     `gorm:"column:yy1_evento_sdh"`
	CreatedAt                 time.Time  `gorm:"column:created_at"`
	UpdatedAt                 time.Time  `gorm:"column:updated_at"`
	InvcID                    *int64     `gorm:"column:invc_id"`
}

func (SalesOrder) TableName() string {
	return "tickets"
}
