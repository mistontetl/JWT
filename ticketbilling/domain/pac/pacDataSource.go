package pac

import (
	pac_service "portal_autofacturacion/domain/pac/service"
	"portal_autofacturacion/models"
)

/*
	type PacDataSource interface {
		StampCFDI( //
			ticketData models.TicketData,
			//preStampedXML []byte, // XML
		) (models.TimbreResponse, error)

		//Other methods!!
	}
*/
func NewPacDataSource(config string) PacDataSource {
	switch config {
	case "SAPIEN":
		return pac_service.SapienPac{}
	default:
	}
	return pac_service.EdicomPac{}
}

type PacDataSource interface {
	StampCFDI(data models.CFDIData) (models.TimbreResponse, error)
}
