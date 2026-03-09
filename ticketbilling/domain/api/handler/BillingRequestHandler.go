package handler

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"time"

	"net/http"
	"portal_autofacturacion/domain/api/controller"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"
	"strings"

	"gorm.io/gorm"
)

/*
	TODO  -----------------  HTTP -----------------
*/

/*
	TODO  -----------------  HTTP -----------------
*/
/////////////correcto

type BillingHistoryHandler struct {
	controller *controller.BillingRequestController
	estadioDB  *gorm.DB
}

func NewTicketHandler(controller *controller.BillingRequestController) *BillingHistoryHandler {
	return &BillingHistoryHandler{controller: controller}
}
func (h *BillingHistoryHandler) SetEstadioDB(estadioDB *gorm.DB) {
	h.estadioDB = estadioDB
}
func (h *BillingHistoryHandler) TrackingBillingRequest(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader(http.StatusOK)
	var tracking models.InvoiceTracking
	if err := json.NewDecoder(r.Body).Decode(&tracking); err != nil {
		http.Error(w, "DTO NOT VALID", http.StatusBadRequest)
		return
	}

	tracking.UUID = strings.TrimSpace(tracking.UUID)
	if tracking.UUID == "" {
		utils.WriteErr(w, "", fmt.Errorf("request_token es obligatorio"), http.StatusBadRequest)
		return
	}

	invoice, err := h.controller.GetInvoiceStatus(tracking.UUID)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		return
	}

	if invoice == nil {
		utils.WriteErr(w, "", fmt.Errorf("Not found"), http.StatusNotFound)
		return
	}

	res := models.InvoiceStatusResponse{
		RequestToken: invoice.RequestToken,
		Ticket:       invoice.UserInputTicket,
		Status:       1,
		//	Status:       invoice.Status.Code,
		//	Description:  invoice.Status.Description,
		Error: invoice.Error,
	}

	utils.WriteAny(w, models.ResponseServerModel[any]{
		Code:     http.StatusOK,
		Datetime: utils.DateTime(),
		Res:      res, //"OK",
	})
	//return
}

///

func (h *BillingHistoryHandler) CreateBillingRequest(
	w http.ResponseWriter,
	r *http.Request,
) {
	log.Println("[HTTP] CreateBillingRequest - request recibido")
	var req models.Payload

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		//http.Error(w, "Invalid DTO", http.StatusBadRequest)
		log.Printf("[HTTP][ERROR] DTO inválido err=%v", err)
		utils.WriteErr(w, "", fmt.Errorf("Invalid DTO"), http.StatusConflict)
		return
	}
	fmt.Println("CAMPOS ::::::::::::::: ::::::::: ", req)
	//  SOLO crea la factura

	token, isNew, key, err := h.controller.CreateInvoice(req)
	if err != nil {
		if strings.Contains(err.Error(), "409 Conflict") {
			utils.WriteErr(w, "", err, http.StatusConflict)
			//	http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(key)
	//	status := http.StatusOK

	if isNew {
		req.UUID = *key
		req.RetryCount = 0
		fmt.Println("::  ", req)
		if err := h.controller.PublishToQueue(r.Context(), req); err != nil {
			utils.WriteErr(w, "", err, http.StatusConflict)
			return
		}
		//	status = http.StatusAccepted // 202
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusAccepted // 202
	}

	log.Printf(
		"[invoice] token=%s ticket=%s new=%v",
		token, req.TicketFolio, isNew,
	)
	///	req.UUID = *key
	//	req.RetryCount = 0
	/*
		err = h.controller.PublishToQueue(r.Context(), req)
		if err != nil {
			utils.WriteErr(w, "", err, http.StatusConflict)
			return
			//	http.Error(w, err.Error(), http.StatusConflict)
		}*/
	utils.WriteAny(w, models.ResponseServerModel[any]{
		Code:     status,
		Datetime: utils.DateTime(),
		Res: map[string]string{
			"request_token": token.RequestToken,
		},
	})
}

/////

func (h *BillingHistoryHandler) VerifyEstadioTicket(w http.ResponseWriter, r *http.Request) {
	if h.estadioDB == nil {
		utils.WriteErr(w, "", fmt.Errorf("conexión ESTADIO no inicializada"), http.StatusInternalServerError)
		return
	}

	var req models.EstadioValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, "", fmt.Errorf("DTO NOT VALID"), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Ticket) == "" {
		utils.WriteErr(w, "", fmt.Errorf("ticket es obligatorio"), http.StatusBadRequest)
		return
	}

	if req.Monto <= 0 {
		utils.WriteErr(w, "", fmt.Errorf("monto inválido"), http.StatusBadRequest)
		return
	}

	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(req.Fecha))
	if err != nil {
		utils.WriteErr(w, "", fmt.Errorf("fecha inválida, usa formato YYYY-MM-DD"), http.StatusBadRequest)
		return
	}

	var salesOrder models.SalesOrder
	margen := 0.01
	err = h.estadioDB.
		Where(
			"source_ticket_id = ? AND DATE(issue_datetime) = ? AND total_amount BETWEEN ? AND ?",
			req.Ticket,
			fecha.Format("2006-01-02"),
			req.Monto-margen,
			req.Monto+margen,
		).
		First(&salesOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusBadRequest)
			utils.WriteAny(w, models.ResponseServerModel[models.EstadioValidationResponse]{
				Code:     http.StatusBadRequest,
				Msg:      "NO",
				Datetime: utils.DateTime(),
				Res:      models.EstadioValidationResponse{E: false},
			})
			return
		}

		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	utils.WriteAny(w, models.ResponseServerModel[models.EstadioValidationResponse]{
		Code:     http.StatusOK,
		Msg:      "OK",
		Datetime: utils.DateTime(),
		Res: models.EstadioValidationResponse{
			E: true,
			T: &salesOrder.OrdrID,
		},
	})
}

func (h *BillingHistoryHandler) CustomerToVerifiedOrder(w http.ResponseWriter, r *http.Request) {
	if h.estadioDB == nil {
		utils.WriteErr(w, "", fmt.Errorf("conexión ESTADIO no inicializada"), http.StatusInternalServerError)
		return
	}

	var req models.CustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, "", fmt.Errorf("DTO NOT VALID"), http.StatusBadRequest)
		return
	}

	if req.TkID <= 0 {
		fmt.Println(":::::::req.tkid ", req.TkID)
		utils.WriteErr(w, "", fmt.Errorf("tk_id es obligatorio"), http.StatusBadRequest)
		return
	}

	rfc := firstNonEmpty(req.Customer.RFC, req.Customer.TaxID)
	razonSocial := firstNonEmpty(req.Customer.RazonSocial, req.Customer.LegalName)
	regimenFiscal := firstNonNilTrim(req.Customer.RegimenFiscal, req.Customer.TaxRegime)
	usoCFDI := strings.TrimSpace(req.Customer.UsoCFDI)
	codigoPostal := firstNonNilTrim(req.Customer.CodigoPostal, req.Customer.PostalCode)
	correo := firstNonNilTrim(req.Customer.Correo, req.Customer.Email)

	if rfc == "" || razonSocial == "" || regimenFiscal == nil || usoCFDI == "" || codigoPostal == nil || correo == nil {
		utils.WriteErr(w, "", fmt.Errorf("rfc, customer.razon_social, customer.regimen_fiscal, customer.uso_cfdi, customer.codigo_postal y customer.correo son obligatorios"), http.StatusBadRequest)
		return
	}

	if !utils.IsValidEmail(*correo) {
		utils.WriteErr(w, "", fmt.Errorf("customer.corr tiene formato inválido"), http.StatusBadRequest)
		return
	}

	var salesOrder models.SalesOrder
	if err := h.estadioDB.Where("tk_id = ?", req.TkID).First(&salesOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println(":::::::req.tkid ", req.TkID)
			utils.WriteErr(w, "", fmt.Errorf("la orden verificada no existe"), http.StatusNotFound)
			return
		}
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	var response models.CustomerResponse
	err := h.estadioDB.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		customerNew := false

		err := tx.Where("tax_id = ?", rfc).First(&customer).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}

			customer = models.Customer{
				TaxID:      rfc,
				LegalName:  razonSocial,
				TaxRegime:  regimenFiscal,
				PostalCode: codigoPostal,
				Email:      correo,
			}
			if err := tx.Create(&customer).Error; err != nil {
				return err
			}
			customerNew = true
		} else {
			updates := map[string]any{
				"legal_name":  razonSocial,
				"tax_regime":  regimenFiscal,
				"postal_code": codigoPostal,
				"email":       correo,
			}
			if err := tx.Model(&customer).Updates(updates).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.SalesOrder{}).Where("tk_id = ?", req.TkID).Update("cust_id", customer.CustID).Error; err != nil {
			return err
		}

		now := time.Now()
		folio := fmt.Sprintf("SIM-%d", req.TkID)
		var invoice models.CustomerInvoice
		err = tx.Where("source_system = ? AND source_ticket_id = ?", salesOrder.SourceSystem, salesOrder.SourceSaleID).First(&invoice).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			invoice = models.CustomerInvoice{
				SourceSystem:   salesOrder.SourceSystem,
				SourceTicketID: salesOrder.SourceSaleID,
				Folio:          &folio,
				InvoicedAt:     &now,
				AttemptedAt:    now,
				IsInvoiced:     true,
			}
			if err := tx.Create(&invoice).Error; err != nil {
				return err
			}
		} else {
			updatesInvoice := map[string]any{
				"folio":         folio,
				"invoiced_at":   now,
				"attempted_at":  now,
				"is_invoiced":   true,
				"error_code":    nil,
				"error_message": nil,
			}
			if err := tx.Model(&invoice).Updates(updatesInvoice).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.SalesOrder{}).Where("tk_id = ?", req.TkID).Update("invc_id", invoice.InvcID).Error; err != nil {
			return err
		}

		response = models.CustomerResponse{
			TkID:        req.TkID,
			CustID:      customer.CustID,
			CustomerNew: customerNew,
		}
		return nil
	})

	if err != nil {
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	utils.WriteAny(w, models.ResponseServerModel[models.CustomerResponse]{
		Code:     http.StatusOK,
		Msg:      "cliente asociado",
		Datetime: utils.DateTime(),
		Res:      response,
	})
}

func (h *BillingHistoryHandler) VerifyBillingStatusx(w http.ResponseWriter, r *http.Request) {
	var req models.InvoiceTracking

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, "", fmt.Errorf("DTO NOT VALID"), http.StatusBadRequest)
		return
	}

	req.UUID = strings.TrimSpace(req.UUID)

	if req.UUID == "" {
		utils.WriteErr(w, "", fmt.Errorf("u obligatorio"), http.StatusBadRequest)
		return
	}

	invoice, err := h.controller.GetInvoiceStatus(req.UUID)

	if err != nil {
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		return
	}

	if invoice == nil {
		utils.WriteErr(w, "", fmt.Errorf("Not found"), http.StatusNotFound)
		return
	}

	if invoice.StatusID != uint(utils.SUCCESS) && invoice.StatusID != uint(utils.ERROR) {
		utils.WriteErr(w, "", fmt.Errorf("status_id debe ser 4 o 6"), http.StatusConflict)
		return
	}

	////
	xmlContent, err := os.ReadFile("C:/Pro/autofacturacion/xmlprueba.xml")
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	zipped, err := zipXML("xmlprueba.xml", xmlContent)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}
	encrypted, err := encryptBytes(zipped, req.UUID)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	utils.WriteAny(w, models.ResponseServerModel[any]{
		Code:     http.StatusOK,
		Datetime: utils.DateTime(),
		Res:      map[string][]int{"bytes": bytesToIntArray(encrypted)},
	})
}

func (h *BillingHistoryHandler) VerifyBillingStatus(w http.ResponseWriter, r *http.Request) {

	var req models.InvoiceTracking

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, "", fmt.Errorf("DTO NOT VALID"), http.StatusBadRequest)
		return
	}

	req.UUID = strings.TrimSpace(req.UUID)

	if req.UUID == "" {
		utils.WriteErr(w, "", fmt.Errorf("u es obligatorio"), http.StatusBadRequest)
		return
	}

	invoice, err := h.controller.GetInvoiceStatus(req.UUID)

	if err != nil {
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		return
	}

	if invoice == nil {
		utils.WriteErr(w, "", fmt.Errorf("Not found"), http.StatusNotFound)
		return
	}

	if invoice.StatusID != uint(utils.SUCCESS) && invoice.StatusID != uint(utils.ERROR) {
		utils.WriteErr(w, "", fmt.Errorf("status_id debe ser 4 o 6"), http.StatusConflict)
		return
	}

	xmlName := "dePrueba.xml"

	xmlContent, err := loadVerifyStatusXML(invoice)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	zipped, err := zipXML(xmlName, xmlContent)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=verify_status.zip")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipped)))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(zipped); err != nil {
		log.Printf("[VerifyBillingStatus][ERROR] writing zip response: %v", err)
	}
}

func encryptBytes(payload []byte, password string) ([]byte, error) {
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	sealed := gcm.Seal(nil, nonce, payload, nil)
	return append(nonce, sealed...), nil
}
func bytesToIntArray(input []byte) []int {
	output := make([]int, len(input))
	for i, b := range input {
		output[i] = int(b)
	}
	return output
}

const defaultVerifyStatusXMLPath = "C:/Pro/autofacturacion/xmlprueba.xml"

func loadVerifyStatusXML(invoice *models.BillingRequest) ([]byte, error) {
	configuredPath := strings.TrimSpace(os.Getenv("VERIFY_STATUS_XML_PATH"))

	if configuredPath == "" {
		configuredPath = defaultVerifyStatusXMLPath
	}

	if configuredPath != "" {
		content, err := os.ReadFile(configuredPath)
		if err != nil {
			return nil, fmt.Errorf("no se pudo leer XML en VERIFY_STATUS_XML_PATH: %w", err)
		}
		return content, nil
	}

	return []byte(fmt.Sprintf("<billingRequest><requestToken>%s</requestToken><statusId>%d</statusId></billingRequest>", invoice.RequestToken, invoice.StatusID)), nil
}

func zipXML(fileName string, payload []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)

	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return nil, err
	}

	if _, err := fileWriter.Write(payload); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstNonNilTrim(values ...*string) *string {
	for _, value := range values {
		if value == nil {
			continue
		}
		trimmed := strings.TrimSpace(*value)
		if trimmed == "" {
			continue
		}
		v := trimmed
		return &v
	}
	return nil
}
