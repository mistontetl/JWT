package service

import (
	"fmt"
	"portal_autofacturacion/models"
)

type EdicomPac struct{}

func (EdicomPac) StampCFDI( //
	ticketData models.TicketData,
	//preStampedXML []byte, // XML
) (models.TimbreResponse, error) {
	return models.TimbreResponse{}, fmt.Errorf("EDICOM!!!")
}
