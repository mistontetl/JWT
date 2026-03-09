package repository

import (
	"log"

	"portal_autofacturacion/models"
)

type ticketRepositoryReadOnly struct{}

func NewTicketRepositoryReadOnly() TicketStoreRepository {
	return &ticketRepositoryReadOnly{}
}

func (r *ticketRepositoryReadOnly) Upsert(t models.Ticket) (models.Ticket, error) {
	log.Println(" READ_ONLY: Ticket.Upsert bloqueado")
	return t, nil
}
