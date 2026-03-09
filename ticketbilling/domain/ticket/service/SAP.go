package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"
	"strconv"
	"time"
)

// SAPPostTicket solo mantiene la configuración de conexión
type SAPPostTicket struct {
	URL string
}

type sapAPIResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Res  dtoTicket `json:"res"`
}

type ticketRequest struct {
	C string `json:"c"`
}

func (s SAPPostTicket) GetValidTicketForBilligS(ticketID string) (models.Ticket, error) {

	client := &http.Client{Timeout: 5 * time.Second}

	payload := ticketRequest{C: ticketID}
	jsonData, _ := json.Marshal(payload)

	fmt.Printf("JSON Body:::::::: %s\n", string(jsonData))

	//  Crear POST
	req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.Ticket{}, fmt.Errorf("error al crear request: %v", err)
	}

	// 3. Configurar Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("rsg-id", strconv.FormatInt(time.Now().UnixMilli(), 10)+"-WEB")
	req.Header.Set("rsg-params", `{"appVersion":1}`)

	resp, err := client.Do(req)

	if err != nil {

		return models.Ticket{}, utils.ErrServerError //fmt.Errorf("error de conexión con SAP: %v", err)
	}

	defer resp.Body.Close()

	//  Validar Status Code
	if resp.StatusCode == http.StatusNotFound { // 404
		return models.Ticket{}, utils.ErrTicketNotFound
	}
	if resp.StatusCode >= 500 {
		return models.Ticket{}, utils.ErrServerError // Error temporal
	}
	if resp.StatusCode != http.StatusOK {
		return models.Ticket{}, fmt.Errorf("API rechazó la solicitud: status %d", resp.StatusCode)
	}
	var apiRes sapAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiRes); err != nil {
		return models.Ticket{}, fmt.Errorf("error al decodificar respuesta JSON: %v", err)
	}

	if apiRes.Res.DocNum == 0 {
		fmt.Printf("el ticket %s no tiene datos (DocNum 0)\n", ticketID)
		return models.Ticket{}, errors.New("TICKET_NOT_FOUND")
	}
	/////
	printData, _ := json.MarshalIndent(apiRes.Res, "", "  ")
	fmt.Println("**************************************************")
	fmt.Println(" DATOS COMPLETOS DESDE SAP (JSON):")
	fmt.Println(string(printData))
	fmt.Println("**************************************************")
	////
	ticketFinal := models.Ticket{

		SourceSaleUUID: fmt.Sprintf("SAP-%d", apiRes.Res.DocEntry),
		SourceSaleID:   ticketID,
		SourceSystem:   "SAP",
		IssrID:         1,
		CustID:         1,
		IssueDatetime:  time.Now(),
		Currency:       "MXN",
		PaymentForm:    "01",
		Subtotal:       apiRes.Res.DocTotal,
		TotalAmount:    apiRes.Res.DocTotal,
		SalesOrderType: "TA",
		//ComentariosSAP: fmt.Sprintf("DocNum: %d, Total: %.2f", apiRes.Res.DocNum, apiRes.Res.DocTotal),
	}

	fmt.Printf(" Datos API::::: %v\n", apiRes.Res.DocumentLines)

	return ticketFinal, nil
}

type dtoTicket struct {
	DocEntry      int               `json:"DocEntry"`
	DocNum        int               `json:"DocNum"`
	DocTotal      float64           `json:"DocTotal"`
	DocDate       string            `json:"DocDate"`
	Series        int               `json:"Series"`
	DocumentLines []dtoDocumentLine `json:"DocumentLines"`
}

type dtoDocumentLine struct {
	ItemCode        string  `json:"ItemCode"`
	ItemDescription string  `json:"ItemDescription"`
	Quantity        float64 `json:"Quantity"`
	LineTotal       float64 `json:"LineTotal"`
	UnitPrice       float64 `json:"UnitPrice"`
	TaxCode         string  `json:"TaxCode"`
	TaxTotal        float64 `json:"TaxTotal"`
}
