package ticket

import (
	"fmt"
	"os"
	ticket_service "portal_autofacturacion/domain/ticket/service"
	"portal_autofacturacion/models"

	"gorm.io/gorm"
)

type TicketDataSource interface {
	//	GetValidTicketForBilligCCO(ticketID string) (models.Ticket, error)
	GetValidTicketForBilligS(ticketID string) (models.Ticket, error)
	//Other methods!!
}

func NewTicketDataSource(config string, dbLocal *gorm.DB) TicketDataSource {
	fmt.Println("entro a ticket datasource")
	if config == "" {
		fmt.Println("ERROR: La variable 'config' llegó vacía a NewTicketDataSource")
	}
	switch config {
	case "SAP":
		fmt.Println("case sap:::  :::: ")

		return ticket_service.SAPPostTicket{
			URL: os.Getenv("API_URL"),
		}

	default:
		fmt.Println("case default")
		//return service.NewSQLTicketAdapter(db)
		return ticket_service.CCOTicket{}
	}
}
