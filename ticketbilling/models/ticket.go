package models

import "time"

type Ticket struct {
	TkID                      int64      `gorm:"column:tk_id;primaryKey;not null"`
	SourceSaleUUID            string     `gorm:"column:source_ticket_uuid;not null"`
	SourceSaleID              string     `gorm:"column:source_ticket_id;not null"`
	SourceSystem              string     `gorm:"column:source_system;not null"`
	IssrID                    int64      `gorm:"column:issr_id;not null"`
	CustID                    uint64     `gorm:"column:cust_id;not null"`
	IssueDatetime             time.Time  `gorm:"column:issue_datetime;not null"`
	Currency                  string     `gorm:"column:currency;not null"`
	PaymentForm               string     `gorm:"column:payment_form"`
	Subtotal                  float64    `gorm:"column:subtotal;type:numeric(15,3);not null"`
	TotalDiscounts            float64    `gorm:"column:total_discounts;type:numeric(15,3)"`
	TotalTaxTransferred       float64    `gorm:"column:total_tax_transferred;type:numeric(15,3)"`
	TotalTaxWithheld          float64    `gorm:"column:total_tax_withheld;type:numeric(15,3)"`
	TotalAmount               float64    `gorm:"column:total_amount;type:numeric(15,3);not null"`
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
	InvcID                    *uint64    `gorm:"column:invc_id"`

	Invoice     *CustomerInvoice `gorm:"foreignKey:InvcID;references:InvcID"`
	TicketLines []TicketLine     `gorm:"foreignKey:TkID;references:TkID"`
	Cliente     *Cliente         `gorm:"-"`
}

func (Ticket) TableName() string {
	return "tickets"
}

/*
type Ticket struct {
	TkID               uint      `gorm:"tk_id;primaryKey;NOT NULL"`
	CancellationStatus string    `gorm:"cancellation_status"`
	ComentariosSAP     string    `gorm:"comentarios_sap"`
	CreateDate         time.Time `gorm:"column:create_date"`
	DateCreatedTicket  time.Time `gorm:"column:date_created_ticket"`
	ErrorSAP           string    `gorm:"column:error_sap"`
	ErrorTimbrado      string    `gorm:"column:error_timbrado"`
	FormaPago          string    `gorm:"column:forma_pago"`

	IdTicket      *string `gorm:"column:id_ticket"`
	IsGlobal      bool    `gorm:"column:is_global"`
	IsSAP         bool    `gorm:"column:is_sap"`
	ObjectKey     *string `gorm:"column:objectkey"`
	Status        string  `gorm:"column:status"`
	SystemGroupID string  `gorm:"column:system_group_id"`
	SystemID      string  `gorm:"column:system_id"`
	TotalAmount   float64 `gorm:"column:total_amount"`

	// FK
	ClienteID *uint    `gorm:"column:cliente_id"`
	Cliente   *Cliente `gorm:"foreignKey:ClienteID;references:ClienteID"`

	InsID   *uint    `gorm:"column:ins_id"`
	Invoice *Invoice `gorm:"foreignKey:InsID;references:InsID"`

	// N:M
	TicketLines []TicketLine `gorm:"foreignKey:TkID"`
	//
}

func (Ticket) TableName() string {
	return "tickets"
}
*/
