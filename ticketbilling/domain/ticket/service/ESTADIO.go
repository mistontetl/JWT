package service

import (
	"errors"
	"fmt"
	"log"

	"portal_autofacturacion/models"

	"gorm.io/gorm"
)

type EstadioTicket struct {
	dbRemote *gorm.DB
	dbLocal  *gorm.DB
}

func NewEstadioTicket(dbRemote, dbLocal *gorm.DB) *EstadioTicket {
	return &EstadioTicket{
		dbRemote: dbRemote,
		dbLocal:  dbLocal,
	}
}

func (e *EstadioTicket) GetValidTicketForBilligS(ticketID string) (models.Ticket, error) {
	fmt.Println(":::::::::estadio", ticketID)
	var so models.SalesOrder

	err := e.dbRemote.
		Where("source_ticket_id = ?", ticketID).
		First(&so).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Ticket{}, errors.New("orden no encontrada en estadio")
		}
		return models.Ticket{}, err
	}

	log.Println("===== SALES ORDER =====")
	log.Printf("tk_id: %d", so.OrdrID)
	log.Printf("source_ticket_uuid: %s", so.SourceSaleUUID)
	log.Printf("source_ticket_id: %s", so.SourceSaleID)
	log.Printf("source_system: %s", so.SourceSystem)
	log.Printf("issr_id: %d", so.IssrID)
	log.Printf("cust_id: %d", so.CustID)
	log.Printf("issue_datetime: %s", so.IssueDatetime)
	log.Printf("currency: %s", so.Currency)
	log.Printf("payment_form: %s", so.PaymentForm)
	log.Printf("subtotal: %.2f", so.Subtotal)
	log.Printf("total_discounts: %.2f", so.TotalDiscounts)
	log.Printf("total_tax_transferred: %.2f", so.TotalTaxTransferred)
	log.Printf("total_tax_withheld: %.2f", so.TotalTaxWithheld)
	log.Printf("total_amount: %.2f", so.TotalAmount)
	log.Printf("sales_order_type: %s", so.SalesOrderType)
	log.Printf("sold_to_party: %s", so.SoldToParty)
	log.Printf("sales_organization: %s", so.SalesOrganization)
	log.Printf("distribution_channel: %s", so.DistributionChannel)
	log.Printf("organization_division: %s", so.OrganizationDivision)
	log.Printf("purchase_order_by_customer: %s", so.PurchaseOrderByCustomer)
	log.Printf("customer_purchase_order_date: %v", so.CustomerPurchaseOrderDate)
	log.Printf("requested_delivery_date: %v", so.RequestedDeliveryDate)
	log.Printf("payment_method_sap: %s", so.PaymentMethodSAP)
	log.Printf("yy1_evento_sdh: %s", so.YY1EventoSDH)
	log.Printf("created_at: %s", so.CreatedAt)
	log.Printf("updated_at: %s", so.UpdatedAt)
	log.Println("=======================")
	// Guardar cache, normalizar campos, etc.
	ticket := models.Ticket{
		TkID:                      so.OrdrID,
		SourceSaleUUID:            so.SourceSaleUUID,
		SourceSaleID:              so.SourceSaleID,
		SourceSystem:              so.SourceSystem,
		IssrID:                    so.IssrID,
		CustID:                    uint64(so.CustID),
		IssueDatetime:             so.IssueDatetime,
		Currency:                  so.Currency,
		PaymentForm:               so.PaymentForm,
		Subtotal:                  so.Subtotal,
		TotalDiscounts:            so.TotalDiscounts,
		TotalTaxTransferred:       so.TotalTaxTransferred,
		TotalTaxWithheld:          so.TotalTaxWithheld,
		TotalAmount:               so.TotalAmount,
		SalesOrderType:            so.SalesOrderType,
		SoldToParty:               so.SoldToParty,
		SalesOrganization:         so.SalesOrganization,
		DistributionChannel:       so.DistributionChannel,
		OrganizationDivision:      so.OrganizationDivision,
		PurchaseOrderByCustomer:   so.PurchaseOrderByCustomer,
		CustomerPurchaseOrderDate: so.CustomerPurchaseOrderDate,
		RequestedDeliveryDate:     so.RequestedDeliveryDate,
		PaymentMethodSAP:          so.PaymentMethodSAP,
		YY1EventoSDH:              so.YY1EventoSDH,
		CreatedAt:                 so.CreatedAt,
		UpdatedAt:                 so.UpdatedAt,
	}
	fmt.Println(ticket.SourceSystem, "  ::: ", ticket.SourceSaleID, "  ::: ", ticket.TotalAmount, "  ::: ", ticket.IssueDatetime, "  ::: ")
	return ticket, nil
}
