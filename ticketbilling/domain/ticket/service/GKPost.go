package service

import (
	"fmt"
	"portal_autofacturacion/models"
)

type GKPostTicket struct{}

func (GKPostTicket) GetValidTicketForBillig(ticketID models.Payload) (models.TicketData, error) {
	return models.TicketData{}, fmt.Errorf("GKPost Not implemented")
}
