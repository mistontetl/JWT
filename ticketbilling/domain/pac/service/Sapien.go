package service

import (
	"portal_autofacturacion/models"
)

type SapienPac struct{}

func (SapienPac) StampCFDI(
	data models.CFDIData,
) (models.TimbreResponse, error) {

	return models.TimbreResponse{
		UUID: "SAPIEN-TEST-UUID",
	}, nil
}

/*
func (SapienPac) StampCFDI( //
	ticketData models.TicketData,
	//preStampedXML []byte, // XML
) (models.TimbreResponse, error) {
	return models.TimbreResponse{}, fmt.Errorf("Sapien!!!")
}
*/
