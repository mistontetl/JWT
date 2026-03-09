package models

import (
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	// Traking
	UUID uuid.UUID `json:"u"`
	//	BillingRequestID int64

	//Retries
	RetryCount int `json:"retry_count"`

	// Data
	TicketFolio string  `json:"f"`
	Total       float64 `json:"t"`
	RFC         string  `json:"r"`
	Email       string  `json:"e"`
}

type InvoiceTracking struct {
	UUID string `json:"u"`
}

type InvoiceStatusResponse struct {
	RequestToken string `json:"r"`
	Ticket       string `json:"t"`
	Status       int    `json:"s"`
	//Description  string  `json:"description"`
	Error *string `json:"error,omitempty"`
}

// //validar

type EstadioValidationRequest struct {
	Monto  float64 `json:"a"`
	Fecha  string  `json:"d"`
	Ticket string  `json:"i"`
}
type EstadioValidationResponse struct {
	E bool   `json:"e"`
	T *int64 `json:"t,omitempty"`
}
type CustomerRequest struct {
	TkID     int64 `json:"ord"`
	Customer struct {
		RFC           string  `json:"r"`
		TaxID         string  `json:"tax_id"`
		RazonSocial   string  `json:"rzs"`
		LegalName     string  `json:"legal_name"`
		RegimenFiscal *string `json:"r_f"`
		TaxRegime     *string `json:"tax_regime"`
		UsoCFDI       string  `json:"c"`
		CodigoPostal  *string `json:"c_p"`
		PostalCode    *string `json:"postal_code"`
		Correo        *string `json:"corr"`
		Email         *string `json:"email"`
	} `json:"cust"`
}

type CustomerResponse struct {
	TkID        int64  `json:"id"`
	CustID      uint64 `json:"cust_id"`
	InvcID      uint64 `json:"invc_id"`
	Invoiced    bool   `json:"is_invoiced"`
	CustomerNew bool   `json:"customer_new"`
}

// //
type ResponseServerModel[T any] struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Res      T      `json:"res,omitempty"`
	Datetime string `json:"dateTime,omitempty"`
}

type TicketData struct {
	ID       string
	SourceID string // (Ej: "CCO", "GK-POST").

	IssueDate   time.Time
	GrossAmount float64
	Subtotal    float64
	TaxRate     float64
	TaxAmount   float64

	InvoiceUUID string
}

type TimbreResponse struct {
	UUID        string
	XMLTimbrado []byte
	XMLPath     string
}
