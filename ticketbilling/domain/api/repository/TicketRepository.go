package repository

import (
	"portal_autofacturacion/domain/ticket/dto"
	"portal_autofacturacion/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func SaveTicketServer(db *gorm.DB, rows []dto.TicketRow) (*models.Ticket, error) {

	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var ticketCreado models.Ticket

	err := db.Transaction(func(tx *gorm.DB) error {
		header := rows[0]
		custID, err := strconv.ParseUint(header.ClientID, 10, 64)
		if err != nil || custID == 0 {
			custID = 1
		}
		ticket := models.Ticket{
			SourceSaleUUID: header.ObjectKey,
			SourceSaleID:   header.IDTicket,
			SourceSystem:   header.SystemID,
			IssrID:         1,
			CustID:         custID,
			IssueDatetime:  header.DateCreated,
			Currency:       "MXN",
			PaymentForm:    header.FormaPago,
			Subtotal:       header.TotalAmount,
			TotalAmount:    header.TotalAmount,
			SalesOrderType: "TA",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		for i, r := range rows {
			now := time.Now()
			discountValue, err := strconv.ParseFloat(r.Descuento, 64)
			var discountAmount *float64
			if err == nil {
				discountAmount = &discountValue
			}
			line := models.TicketLine{
				TkID:           ticket.TkID,
				LineNumber:     i + 1,
				Description:    r.Descripcion,
				Quantity:       r.Cantidad,
				UnitPrice:      r.ValorUnitario,
				LineAmount:     r.Base,
				DiscountAmount: discountAmount,
				ProductSATCode: r.ClaveProdServ,
				UnitSATCode:    r.ClaveUnidad,
				SKU:            &r.NoIdentificacion,
				CreatedAt:      &now,

				TaxRate:         r.Taxrate,
				TaxRateTypeCode: r.TaxrateTypeCode,
			}

			if err := tx.Create(&line).Error; err != nil {
				return err
			}

		}
		ticketCreado = ticket
		return nil

	})

	if err != nil {
		return nil, err
	}

	return &ticketCreado, nil
}
