package service

import (
	"fmt"
	"log"
	"portal_autofacturacion/domain/api/repository"
	"portal_autofacturacion/domain/ticket/dto"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"

	"gorm.io/gorm"
)

type CCOTicket struct {
	dbCCO   *gorm.DB
	dbLocal *gorm.DB
}

func NewCCOTicket(dbCCO *gorm.DB, dbLocal *gorm.DB) CCOTicket {
	fmt.Println("New CCOTicker")

	return CCOTicket{
		dbCCO:   dbCCO,
		dbLocal: dbLocal,
	}
}

/*
	func (CCOTicket) GetValidTicketForBillig(ticketID models.Payload) (models.TicketData, error) {
		return models.TicketData{}, fmt.Errorf("CCO!!!")
	}
*/
/*
func (c CCOTicket) GetValidTicketForBilligS(ticketID string) (models.Ticket, error) {
	return models.Ticket{}, nil
}
*/
func (c CCOTicket) GetValidTicketForBilligS(ticketID string) (models.Ticket, error) {
	fmt.Println("GetValidTicketForBilligS::::::::::::", ticketID)
	if ticketID == "" {
		log.Println("ticketID vacío, usando ID de prueba")
		ticketID = "POS1100001081"
	}

	rows, err := TicketConsulta(c.dbCCO, ticketID)
	if err != nil {
		return models.Ticket{}, err
	}

	if len(rows) == 0 {
		return models.Ticket{}, utils.ErrTicketNotFound
	}

	fmt.Println("===================================")
	fmt.Println("REGISTROS OBTENIDOS DESDE CCO")
	fmt.Println("===================================")

	for _, r := range rows {
		fmt.Printf(
			r.ObjectKey, r.SystemID, r.IDTicket, r.SystemGroupID, r.TcklID, r.IDTicketBP, r.DateCreated, r.TotalAmount, r.Cantidad, r.Amount, r.NoIdentificacion, r.Descuento, r.PorcentajeDescuento, r.TaxrateTypeCode, r.Taxrate, r.ValorUnitario, r.ClientID, r.Status, r.CancellationStatus, r.Descripcion, r.RFC, r.CodigoPostal, r.RegimenFiscal, r.CombinedName, r.ClaveUnidad, r.ClaveProdServ, r.Base, r.FormaPago,
		)
	}

	ticket, err := repository.SaveTicketServer(c.dbLocal, rows)
	if err != nil {
		return models.Ticket{}, err
	}

	log.Println("Ticket guardado con TkID:", ticket.TkID)

	return *ticket, nil

	/*
		id := "40030002"
		return models.Ticket{
			//	IdTicket:    &ticketID,
			IdTicket:    &id,
			IsSAP:       true,
			TotalAmount: 1000.00,
			Status:      rows[0].Status,
		}, nil*/
}

func TicketConsulta(db *gorm.DB, ticketID string) ([]dto.TicketRow, error) {
	sql := `
WITH payment_rank AS (
    SELECT
        pi.REC_KEY,
        pi.CREDITCARDTYPECODE,
        SUM(pi.BUSINESSTRANSACTIONAMOUNT) AS TOTAL_AMOUNT,
        ROW_NUMBER() OVER (
            PARTITION BY pi.REC_KEY
            ORDER BY 
                SUM(pi.BUSINESSTRANSACTIONAMOUNT) DESC,
                pi.CREDITCARDTYPECODE ASC
        ) AS rn
    FROM CCOM_PRUEBAS.PAYMENTITEM pi
    GROUP BY pi.REC_KEY, pi.CREDITCARDTYPECODE
)
SELECT
    re.OBJECTKEY                AS objectkey,
    re.POS_SYSTEM_NAME          AS systemId,
    re.ID                       AS idTicket,
    re.POS_GROUP_ID             AS systemGroupId,
    it.OBJECTKEY                AS tckl_id,
    bp.OBJECTKEY                AS idTicket_bp,
    re.BUSINESSTRANSACTIONDATE  AS dateCreated,
    re.PAYMENTGROSSAMOUNT       AS totalAmount,
    it.QUANTITY                 AS cantidad,
    it.UNITGROSSAMOUNTORIGIN    AS amount,
    it.ID                       AS noIdentificacion,
    it.DISCOUNT_NET_AMOUNT      AS descuento,
    it.DISCOUNTPERCENTAGE       AS porcentajeDescuento,
    it.TAXRATETYPECODE          AS taxrateTypeCode,
    it.TAXRATE                  AS taxrate,
    it.UNITGROSSAMOUNT          AS valorUnitario,
    re.BUPAEXTERNALID           AS client_id,
    re.STATUS                   AS status,
    re.CANCELLATIONSTATUS       AS cancellationStatus,
    CASE
        WHEN mt.PRODCATID = '102' THEN 'MEZCLA DE SUSTANCIAS ODORIFERAS'
        ELSE it.DESCRIPTION
    END AS descripcion,
    bp.TAX_ID                   AS rfc,
    bp.POSTALCODE               AS codigoPostal,
    bp.FAXNUMBER                AS regimenFiscal,
    bp.COMBINEDNAME             AS combinedname,
    mt.DEFAULTQUANTITYTYPECODE  AS claveUnidad,
    mt.UDF_STRING_XL3           AS claveProdServ,
    it.PAYMENTNETAMOUNT         AS base,
    CASE
        WHEN pg.CREDITCARDTYPECODE IS NULL THEN '01'
        ELSE pg.CREDITCARDTYPECODE
    END AS formaPago
FROM CCOM_PRUEBAS.RECEIPT re
INNER JOIN CCOM_PRUEBAS.SALESITEM it
    ON re.OBJECTKEY = it.RECEIPT_OBJECTKEY
LEFT JOIN CCOM_PRUEBAS.BUSINESSPARTNER bp
    ON bp.EXTERNALID = re.BUPAEXTERNALID
LEFT JOIN CCOM_PRUEBAS.MATERIAL mt
    ON it.MAT_KEY = mt.OBJECTKEY
LEFT JOIN payment_rank pg
    ON pg.REC_KEY = re.OBJECTKEY
   AND pg.rn = 1
WHERE re.TOTALGROSSAMOUNT > 0
  AND re.ID = ?
  AND re.TYPECODE = 1
  AND re.CANCELLATIONSTATUS != 4
  AND re.IS_REVERSE_RECEIPT != 1
  AND it.STATUS = 2
ORDER BY re.BUSINESSTRANSACTIONDATE DESC;
`

	fmt.Println("Parametro TicketID:", ticketID)
	var receipts []dto.TicketRow

	err := db.Debug().Raw(sql, ticketID).Scan(&receipts).Error

	if err != nil {
		return nil, err
	}

	return receipts, nil
}
