package repository

import "portal_autofacturacion/models"

type TicketStoreRepository interface {
	Upsert(t models.Ticket) (models.Ticket, error)
}
