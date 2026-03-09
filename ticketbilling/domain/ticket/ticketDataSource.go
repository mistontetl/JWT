package ticket

import (
	"fmt"
	"os"
	conexion "portal_autofacturacion/conexiones"
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
	case "ESTADIO":
		fmt.Println("case Estadio:::  :::: ")
		dbSQLSt, err := conexion.ConexionEstadio()
		if err != nil {
			panic("error conectando a estadio: " + err.Error())
		}
		return ticket_service.NewEstadioTicket(dbSQLSt, dbLocal)
	default:
		fmt.Println("case default")
		//return service.NewSQLTicketAdapter(db)
		return ticket_service.CCOTicket{}
	}
}
