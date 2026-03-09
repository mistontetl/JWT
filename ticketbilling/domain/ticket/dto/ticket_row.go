package dto

import "time"

type TicketRow struct {
	ObjectKey           string
	SystemID            string
	IDTicket            string
	SystemGroupID       string
	TcklID              string
	IDTicketBP          string
	DateCreated         time.Time
	TotalAmount         float64
	Cantidad            float64
	Amount              float64
	NoIdentificacion    string
	Descuento           string
	PorcentajeDescuento string
	TaxrateTypeCode     string
	Taxrate             float64
	ValorUnitario       float64
	ClientID            string
	Status              string
	CancellationStatus  int
	Descripcion         string
	RFC                 string
	CodigoPostal        string
	RegimenFiscal       string
	CombinedName        string
	ClaveUnidad         string
	ClaveProdServ       string
	Base                float64
	FormaPago           string
}
