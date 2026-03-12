package service

import (
	"fmt"
	"portal_autofacturacion/models"
)

type SapienPac struct{}

func (SapienPac) StampCFDI( //
	ticketData models.TicketData,
	//preStampedXML []byte, // XML
) (models.TimbreResponse, error) {
	return models.TimbreResponse{}, fmt.Errorf("Sapien!!!")
}
