package models

type CFDIData struct {
	Ticket  Ticket
	Cliente Cliente
	Lineas  []TicketLine
}
