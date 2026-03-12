package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"

	"portal_autofacturacion/config"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/domain/api/repository"
	pac_service "portal_autofacturacion/domain/pac"
	ticket_service "portal_autofacturacion/domain/ticket"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"

	"gorm.io/gorm"
)

const (
	MaxConcurrentJobs = 3
	MaxRetries        = 5
	TOLERANCIA        = 0.01
)

type invoiceWorker struct {
	BillingRepo      repository.BillingRequestRepository
	TicketRepo       repository.TicketStoreRepository
	Conexion         *gorm.DB
	sem              chan struct{}
	client           queue.RSGQueue
	TicketDataSource ticket_service.TicketDataSource
	PacDataSource    pac_service.PacDataSource
}

func NewInvoiceWorker(
	client queue.RSGQueue,
	runtimeConfig config.WorkerRuntimeConfig,
	db *gorm.DB,
) invoiceWorker {
	conexion := db
	mode := runtimeConfig.Mode

	var billingRepo repository.BillingRequestRepository
	var ticketRepo repository.TicketStoreRepository

	if mode == "READ_ONLY" {
		log.Println(" WORKER en modo READ_ONLY 000000")
		billingRepo = repository.NewBillingRequestRepositoryReadOnly()
		ticketRepo = repository.NewTicketRepositoryReadOnly()
	} else {
		billingRepo = repository.NewBillingRequestRepositoryPru(conexion)
		ticketRepo = repository.NewTicketRepositoryDB(conexion)
	}

	return invoiceWorker{
		BillingRepo:      billingRepo,
		TicketRepo:       ticketRepo,
		Conexion:         conexion,
		sem:              make(chan struct{}, MaxConcurrentJobs),
		client:           client,
		TicketDataSource: ticket_service.NewTicketDataSource(runtimeConfig.TicketSource, conexion),
		PacDataSource:    pac_service.NewPacDataSource(runtimeConfig.PacProvider),
	}
}

func (w *invoiceWorker) HandleDelivery(delivery queue.Delivery) {
	w.sem <- struct{}{}
	defer func() { <-w.sem }()

	log.Println("[WORKER] Inicio HandleDelivery 0000000")

	var payload models.Payload
	if err := json.Unmarshal(delivery.Body(), &payload); err != nil {
		log.Println("Payload inválido:", err)
		delivery.Ack()
		return
	}

	log.Printf(
		"Payload: UUID=%s Folio=%s Monto=%.2f Intento=%d",
		payload.UUID,
		payload.TicketFolio,
		payload.Total,
		payload.RetryCount,
	)

	ok, err := w.BillingRepo.Claim(payload.UUID.String())
	if err != nil || !ok {
		log.Println("BillingRequest no reclamable (repo)")
		delivery.Ack()
		return
	}

	ticketDB, err := w.findTicketCache(payload.TicketFolio)
	if err != nil {
		log.Println("Error BD consultando ticket cache:", err)
		w.retry(payload, delivery)
		return
	}

	if ticketDB != nil {
		if math.Abs(ticketDB.TotalAmount-payload.Total) > TOLERANCIA {
			errMsg := "Monto no coincide con el registro"
			w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		w.linkTicket(payload.UUID.String(), uint(ticketDB.TkID))
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.STAMPING, nil)

		if err := w.processXML(payload, ticketDB); err != nil {
			delivery.Ack()
			return
		}

		log.Printf("SUCCESS [%s] Ticket %s", payload.UUID, payload.TicketFolio)
		delivery.Ack()
		return
	}

	log.Println("Consultando fuente externa:", payload.TicketFolio)
	extTicket, err := w.TicketDataSource.GetValidTicketForBilligS(payload.TicketFolio)
	if err != nil {
		if errors.Is(err, utils.ErrTicketNotFound) {
			errMsg := "Ticket no encontrado en fuente externa"
			w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		if utils.IsTemporaryError(err) || errors.Is(err, utils.ErrServerError) {
			w.retry(payload, delivery)
			return
		}

		errMsg := err.Error()
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	extTicket.SourceSaleID = payload.TicketFolio
	if extTicket.SourceSystem == "" {
		extTicket.SourceSystem = ""
	}

	ticketUpsert, err := w.TicketRepo.Upsert(extTicket)
	if err != nil {
		w.retry(payload, delivery)
		return
	}

	if math.Abs(ticketUpsert.TotalAmount-payload.Total) > TOLERANCIA {
		errMsg := "Monto no coincide con el registro externo"
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	ticketForXML, err := w.loadTicketForXML(ticketUpsert)
	if err != nil {
		w.retry(payload, delivery)
		return
	}

	//	w.linkTicket(payload.UUID.String(), ticketForXML.TkID)
	//
	w.linkTicket(payload.UUID.String(), uint(ticketForXML.TkID))
	w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.STAMPING, nil)
	if err := w.processXML(payload, &ticketForXML); err != nil {
		delivery.Ack()
		return
	}

	log.Printf("SUCCESS [%s] Ticket %s", payload.UUID, payload.TicketFolio)
	delivery.Ack()
}

func (w *invoiceWorker) StartConsuming(consumer queue.Consumer) error {
	log.Println("StartConsuming llamado, esperando mensajes...")
	return consumer.ConsumeAsync(w.HandleDelivery)
}

func (w *invoiceWorker) retry(payload models.Payload, delivery queue.Delivery) {
	if os.Getenv("WORKER_MODE") == "READ_ONLY" {
		log.Println(" READ_ONLY: retry deshabilitado")
		delivery.Ack()
		return
	}

	if w.Conexion == nil {
		payload.RetryCount++
		delay := int(math.Pow(2, float64(payload.RetryCount))) * 30
		w.client.PublishWithRetry(context.Background(), payload, delay)
		delivery.Ack()
		return
	}

	var br models.BillingRequest
	err := w.Conexion.Where("request_token = ?", payload.UUID.String()).First(&br).Error
	if err != nil {
		payload.RetryCount++
		delay := int(math.Pow(2, float64(payload.RetryCount))) * 30
		w.client.PublishWithRetry(context.Background(), payload, delay)
		delivery.Ack()
		return
	}

	if br.BrRetryCount >= MaxRetries {
		errMsg := "Máximo de reintentos alcanzado"
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	newRetry := br.BrRetryCount + 1
	delay := int(math.Pow(2, float64(newRetry))) * 30
	w.updateStatusWithRetry(payload.UUID.String(), utils.PENDING, nil, newRetry)

	payload.RetryCount = newRetry
	w.client.PublishWithRetry(context.Background(), payload, delay)
	delivery.Ack()
}

func (w *invoiceWorker) findTicketCache(ticketFolio string) (*models.Ticket, error) {
	if w.Conexion == nil {
		return nil, nil
	}

	var ticketDB models.Ticket
	err := w.Conexion.
		Preload("Cliente").
		Preload("TicketLines").
		Where("id_ticket = ? AND is_sap = true", ticketFolio).
		First(&ticketDB).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &ticketDB, nil
}

func (w *invoiceWorker) loadTicketForXML(ticket models.Ticket) (models.Ticket, error) {
	if w.Conexion == nil {
		return ticket, nil
	}

	if ticket.TkID == 0 {
		return ticket, nil
	}

	err := w.Conexion.
		Preload("Cliente").
		Preload("TicketLines").
		Where("tk_id = ?", ticket.TkID).
		First(&ticket).Error
	if err != nil {
		return models.Ticket{}, err
	}

	return ticket, nil
}

func (w *invoiceWorker) linkTicket(token string, tkID uint) {
	if w.Conexion == nil || os.Getenv("WORKER_MODE") == "READ_ONLY" {
		return
	}

	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Update("ticket_id", tkID)
}

func (w *invoiceWorker) updateStatusWithRetry(token string, status int, errMsg *string, retryCount int) {
	if w.Conexion == nil || os.Getenv("WORKER_MODE") == "READ_ONLY" {
		return
	}

	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":      status,
			"error":          errMsg,
			"br_retry_count": retryCount,
		})
}

func (w *invoiceWorker) genericClient() (*models.Cliente, error) {
	if w.Conexion == nil {
		return nil, gorm.ErrInvalidDB
	}

	if os.Getenv("WORKER_MODE") == "READ_ONLY" {
		cliente := models.Cliente{
			Nombre:             "PUBLICO EN GENERAL",
			Email:              "facturacion@empresa.com",
			RFC:                "XAXX010101000",
			RegimenFiscal:      "616",
			DescripcionRegimen: "Sin obligaciones fiscales",
			PostalCode:         "00000",
			ExternalID:         "PUBLICO_GENERAL",
		}
		return &cliente, nil
	}

	var cliente models.Cliente
	err := w.Conexion.
		Where("rfc = ?", "XAXX010101000").
		First(&cliente).Error

	if err == nil {
		return &cliente, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cliente = models.Cliente{
		Nombre:             "PUBLICO EN GENERAL",
		Email:              "facturacion@empresa.com",
		RFC:                "XAXX010101000",
		RegimenFiscal:      "616",
		DescripcionRegimen: "Sin obligaciones fiscales",
		PostalCode:         "00000",
		ExternalID:         "PUBLICO_GENERAL",
	}

	if err := w.Conexion.Create(&cliente).Error; err != nil {
		return nil, err
	}

	return &cliente, nil
}

func (w *invoiceWorker) processXML(payload models.Payload, ticketDB *models.Ticket) error {
	if ticketDB.TkID == 0 {
		errMsg := "Ticket inválido"
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		return errors.New(errMsg)
	}

	if ticketDB.Cliente == nil {
		publicClient, err := w.genericClient()
		if err != nil {
			errMsg := "No se pudo obtener cliente genérico"
			w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			return err
		}

		if w.Conexion != nil && os.Getenv("WORKER_MODE") != "READ_ONLY" {
			if err := w.Conexion.Model(ticketDB).
				Update("cliente_id", publicClient.ClienteID).Error; err != nil {
				errMsg := "Error al asociar cliente público"
				w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
				return err
			}
		}
		ticketDB.Cliente = publicClient
	}

	if len(ticketDB.TicketLines) == 0 {
		errMsg := "Ticket sin conceptos"
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		return errors.New(errMsg)
	}

	cfdiData := models.CFDIData{
		Ticket:  *ticketDB,
		Cliente: *ticketDB.Cliente,
		Lineas:  ticketDB.TicketLines,
	}

	resp, err := w.PacDataSource.StampCFDI(cfdiData)
	if err != nil {
		errMsg := err.Error()
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		return err
	}

	log.Println("CFDI timbrado correctamente")
	log.Println("UUID CFDI:", resp.UUID)
	log.Println("XML PATH:", resp.XMLPath)

	w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.SUCCESS, nil)
	return nil
}

/*package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"portal_autofacturacion/config"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/domain/api/repository"
	pac_service "portal_autofacturacion/domain/pac"
	ticket_service "portal_autofacturacion/domain/ticket"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"














	"gorm.io/gorm"
)

const (
	MaxConcurrentJobs = 3
	MaxRetries        = 5
	TOLERANCIA        = 0.01
)

type invoiceWorker struct {
	BillingRepo      repository.BillingRequestRepository //
	Conexion         *gorm.DB
	sem              chan struct{}
	client           queue.RSGQueue
	TicketDataSource ticket_service.TicketDataSource
	PacDataSource    pac_service.PacDataSource
}

func NewInvoiceWorker(client queue.RSGQueue, ticketConfig string, pacConfig string, conexion *gorm.DB) invoiceWorker {
	mode := os.Getenv("WORKER_MODE")

	var repo repository.BillingRequestRepository
	if mode == "READ_ONLY" {
		fmt.Println("solo lectura")
		repo = repository.NewBillingRequestRepositoryReadOnly()
	} else {
		fmt.Println("No lectura")
		repo = repository.NewBillingRequestRepositoryPru(conexion)
	}

	return invoiceWorker{
		BillingRepo:      repo,
		Conexion:         conexion,
		sem:              make(chan struct{}, MaxConcurrentJobs),
		client:           client,
		TicketDataSource: ticket_service.NewTicketDataSource(ticketConfig, conexion),
		PacDataSource:    pac_service.NewPacDataSource(pacConfig),
	}
}

func (w *invoiceWorker) HandleDelivery(delivery queue.Delivery) {
	w.sem <- struct{}{}
	defer func() { <-w.sem }()
	log.Println("Inicio HandleDelivery")
	log.Println("[WORKER] Inicio HandleDelivery")

	var payload models.Payload

	if err := json.Unmarshal(delivery.Body(), &payload); err != nil {
		log.Println("Payload inválido:", err)
		delivery.Ack()
		return
	}

	log.Printf("Payload: UUID=%s Folio=%s Monto=%.2f Intento=%d",
		payload.UUID,
		payload.TicketFolio,
		payload.Total,
		payload.RetryCount,
	)

	//  Reclamar tarea
	fmt.Println(payload.UUID.String())

	res := w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ? AND status_id = ?", payload.UUID.String(), utils.PENDING).
		Updates(map[string]interface{}{
			"status_id":  utils.VALIDATING,
			"updated_at": time.Now(),
		})

	fmt.Println("respuesta ::: ", res.RowsAffected)

	if res.Error != nil || res.RowsAffected == 0 {
		log.Println("BillingRequest no reclamable o ya tomado")
		delivery.Ack()
		return
	}

	//  Buscar ticket en cache
	var ticketDB models.Ticket
	fmt.Println("::::::::::ID DEL TICKET", payload.TicketFolio)
	err := w.Conexion.
		Preload("Cliente").
		Preload("TicketLines").
		Where("id_ticket = ? AND is_sap = true", payload.TicketFolio).
		First(&ticketDB).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error BD:", err)
		w.retry(payload, delivery)
		return
	}

	if err == nil {
		log.Println("Ticket encontrado en BD")

		if math.Abs(ticketDB.TotalAmount-payload.Total) > TOLERANCIA {
			log.Printf(" FRAUDE: DB=%.2f Usuario=%.2f", ticketDB.TotalAmount, payload.Total)
			errMsg := "Monto no coincide con el registro"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		w.linkTicket(payload.UUID.String(), ticketDB.TkID)

		w.finishSuccess(payload.UUID.String())
		delivery.Ack()

		w.processXML(payload, &ticketDB, delivery)
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("DB error leyendo ticket: %v", err)
		w.retry(payload, delivery)
		return
	}

	log.Println(" NO se encontró ticket en BD local, consultando CCO")
	// Fuente externa
	log.Println("Consultando fuente externa")
	extTicket, err := w.TicketDataSource.GetValidTicketForBilligS(payload.TicketFolio)

	if err != nil {

		if errors.Is(err, utils.ErrTicketNotFound) {
			log.Printf("Ticket no encontrado (404) para Folio: %s", payload.TicketFolio)
			errMsg := "Ticket no encontrado en fuente externa"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		if utils.IsTemporaryError(err) || errors.Is(err, utils.ErrServerError) {
			w.retry(payload, delivery)
			return
		}

		errMsg := err.Error()
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	//  UPSERT

	extTicket.IdTicket = &payload.TicketFolio
	extTicket.IsSAP = true

	ticketDB, err = w.upsertTicket(extTicket)
	if err != nil {
		w.retry(payload, delivery)
		return
	}

	//  post-fetch
	if math.Abs(ticketDB.TotalAmount-payload.Total) > TOLERANCIA {
		log.Printf(" (EXTERNO): Externo=%.2f Usuario=%.2f", ticketDB.TotalAmount, payload.Total)

		errMsg := "Monto no coincide con el registro externo"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	w.linkTicket(payload.UUID.String(), ticketDB.TkID)

	///
	w.updateStatus(payload.UUID.String(), utils.STAMPING, nil)

	time.Sleep(5 * time.Second)

	w.finishSuccess(payload.UUID.String())

	log.Printf("SUCCESS [%s] Ticket %s", payload.UUID, payload.TicketFolio)
	delivery.Ack()
}

func (w *invoiceWorker) StartConsuming(consumer queue.Consumer) error {
	log.Println(" StartConsuming llamado, esperando mensajes...")
	return consumer.ConsumeAsync(w.HandleDelivery)
}
func (w *invoiceWorker) retry(payload models.Payload, delivery queue.Delivery) {
	var br models.BillingRequest
	w.Conexion.Where("request_token = ?", payload.UUID.String()).First(&br)

	if br.BrRetryCount >= MaxRetries {
		errMsg := "Máximo de reintentos alcanzado"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	newRetry := br.BrRetryCount + 1
	delay := int(math.Pow(2, float64(newRetry))) * 30

	w.updateStatusWithRetry(payload.UUID.String(), utils.PENDING, nil, newRetry)

	payload.RetryCount = newRetry
	w.client.PublishWithRetry(context.Background(), payload, delay)

	delivery.Ack()
}

func (w *invoiceWorker) upsertTicket(t models.Ticket) (models.Ticket, error) {
	var dbT models.Ticket

	err := w.Conexion.
		Where("id_ticket = ? AND is_sap = true", *t.IdTicket).
		First(&dbT).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return t, w.Conexion.Create(&t).Error
	}

	if err != nil {
		return models.Ticket{}, err
	}

	err = w.Conexion.Model(&dbT).
		Updates(map[string]interface{}{
			"total_amount": t.TotalAmount,
			"status":       t.Status,
			"updated_at":   time.Now(),
		}).Error

	return dbT, err
}

func (w *invoiceWorker) linkTicket(token string, tkID uint) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Update("ticket_id", tkID)
}

func (w *invoiceWorker) finishSuccess(token string) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":  utils.SUCCESS,
			"updated_at": time.Now(),
		})
}

func (w *invoiceWorker) updateStatus(token string, status int, errMsg *string) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":  status,
			"error":      errMsg,
			"updated_at": time.Now(),
		})
}

func (w *invoiceWorker) updateStatusWithRetry(token string, status int, errMsg *string, retryCount int) {
	w.Conexion.Model(&models.BillingRequest{}).
		Where("request_token = ?", token).
		Updates(map[string]interface{}{
			"status_id":      status,
			"error":          errMsg,
			"br_retry_count": retryCount,
			"updated_at":     time.Now(),
		})
}

/////Crea cliente

func (w *invoiceWorker) genericClient() (*models.Cliente, error) {
	var cliente models.Cliente

	err := w.Conexion.
		Where("rfc = ?", "XAXX010101000").
		First(&cliente).Error

	if err == nil {
		return &cliente, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cliente = models.Cliente{
		Nombre:             "PUBLICO EN GENERAL",
		Email:              "facturacion@empresa.com",
		RFC:                "XAXX010101000",
		RegimenFiscal:      "616",
		DescripcionRegimen: "Sin obligaciones fiscales",
		PostalCode:         "00000",
		ExternalID:         "PUBLICO_GENERAL",
	}

	if err := w.Conexion.Create(&cliente).Error; err != nil {
		return nil, err
	}

	return &cliente, nil
}

// PROCESO XML
// ======================================================
func (w *invoiceWorker) processXML(
	payload models.Payload,
	ticketDB *models.Ticket,
	delivery queue.Delivery,
) {

	// Validaciones base
	fmt.Println("validaciones de xml")
	if ticketDB.TkID == 0 {
		errMsg := "Ticket inválido"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	// Cliente público
	if ticketDB.Cliente == nil {
		log.Println("Asignando cliente PUBLICO EN GENERAL")

		publicClient, err := w.genericClient()
		if err != nil {
			errMsg := "No se pudo obtener cliente genérico"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		if err := w.Conexion.Model(ticketDB).
			Update("cliente_id", publicClient.ClienteID).Error; err != nil {
			errMsg := "Error al asociar cliente público"
			w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}
		// Refrescar relaciones en memoria
		ticketDB.Cliente = publicClient
	}

	if len(ticketDB.TicketLines) == 0 {
		errMsg := "Ticket sin conceptos"
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	// Vincular ticket
	fmt.Println("entra a xml")
	w.linkTicket(payload.UUID.String(), ticketDB.TkID)

	// Armar CFDI
	cfdiData := models.CFDIData{
		Ticket:  *ticketDB,
		Cliente: *ticketDB.Cliente,
		Lineas:  ticketDB.TicketLines,
	}

	resp, err := w.PacDataSource.StampCFDI(cfdiData)
	if err != nil {
		errMsg := err.Error()
		w.updateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	log.Println("CFDI timbrado correctamente")
	log.Println("UUID CFDI:", resp.UUID)
	log.Println("XML PATH:", resp.XMLPath)

	w.finishSuccess(payload.UUID.String())
	delivery.Ack()
}
*/
/*
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"

	"gorm.io/gorm"
	"portal_autofacturacion/config"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/domain/api/repository"
	pac_service "portal_autofacturacion/domain/pac"
	ticket_service "portal_autofacturacion/domain/ticket"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"
)

const (
	MaxConcurrentJobs = 3
	MaxRetries        = 5
	TOLERANCIA        = 0.01
)

type invoiceWorker struct {
	BillingRepo      repository.BillingRequestRepository
	TicketRepo       repository.TicketStoreRepository
	Conexion         *gorm.DB
	sem              chan struct{}
	client           queue.RSGQueue
	TicketDataSource ticket_service.TicketDataSource
	PacDataSource    pac_service.PacDataSource
}

func NewInvoiceWorker(
	client queue.RSGQueue,
	ticketConfig string,
	pacConfig string,
	db any, //  solo para construir datasources, NO se usa directo
) invoiceWorker {

	mode := os.Getenv("WORKER_MODE")

	var billingRepo repository.BillingRequestRepository
	var ticketRepo repository.TicketStoreRepository

	if mode == "READ_ONLY" {
		log.Println(" WORKER en modo READ_ONLY 000000")
		billingRepo = repository.NewBillingRequestRepositoryReadOnly()
		ticketRepo = repository.NewTicketRepositoryReadOnly()

	} else {
		billingRepo = repository.NewBillingRequestRepositoryPru(db.(*gorm.DB))
		ticketRepo = repository.NewTicketRepositoryDB(db.(*gorm.DB))
	}

	return invoiceWorker{
		BillingRepo:      billingRepo,
		TicketRepo:       ticketRepo,
		sem:              make(chan struct{}, MaxConcurrentJobs),
		client:           client,
		TicketDataSource: ticket_service.NewTicketDataSource(ticketConfig, db.(*gorm.DB)),
		PacDataSource:    pac_service.NewPacDataSource(pacConfig),
	}
}

func (w *invoiceWorker) HandleDelivery(delivery queue.Delivery) {
	w.sem <- struct{}{}
	defer func() { <-w.sem }()

	log.Println("[WORKER] Inicio HandleDelivery 0000000")

	var payload models.Payload
	if err := json.Unmarshal(delivery.Body(), &payload); err != nil {
		log.Println("Payload inválido:", err)
		delivery.Ack()
		return
	}

	log.Printf(
		"Payload: UUID=%s Folio=%s Monto=%.2f Intento=%d",
		payload.UUID,
		payload.TicketFolio,
		payload.Total,
		payload.RetryCount,
	)

	//  Reclamo vía repositorio (READ_ONLY-safe)
	ok, err := w.BillingRepo.Claim(payload.UUID.String())
	if err != nil || !ok {
		log.Println("BillingRequest no reclamable (repo)")
		delivery.Ack()
		return
	}

	//  Consultar fuente externa (ESTADIO / CCO / SAP)
	log.Println("Consultando fuente externa:", payload.TicketFolio)

	extTicket, err := w.TicketDataSource.GetValidTicketForBilligS(payload.TicketFolio)
	if err != nil {

		if errors.Is(err, utils.ErrTicketNotFound) {
			errMsg := "Ticket no encontrado en fuente externa"
			w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
			delivery.Ack()
			return
		}

		if utils.IsTemporaryError(err) || errors.Is(err, utils.ErrServerError) {
			w.retry(payload, delivery)
			return
		}

		errMsg := err.Error()
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	// Normalización mínima
	extTicket.IdTicket = &payload.TicketFolio
	extTicket.IsSAP = true

	//  Validar monto
	if math.Abs(extTicket.TotalAmount-payload.Total) > TOLERANCIA {
		errMsg := "Monto no coincide con el registro externo"
		w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.ERROR, &errMsg)
		delivery.Ack()
		return
	}

	//  Upsert (bloqueado en READ_ONLY)
	_, err = w.TicketRepo.Upsert(extTicket)
	if err != nil {
		w.retry(payload, delivery)
		return
	}

	//  Flujo OK
	w.BillingRepo.UpdateStatus(payload.UUID.String(), utils.SUCCESS, nil)
	log.Printf("SUCCESS [%s] Ticket %s", payload.UUID, payload.TicketFolio)

	delivery.Ack()
}

func (w *invoiceWorker) StartConsuming(consumer queue.Consumer) error {
	log.Println("StartConsuming llamado, esperando mensajes...")
	return consumer.ConsumeAsync(w.HandleDelivery)
}

func (w *invoiceWorker) retry(payload models.Payload, delivery queue.Delivery) {
	if os.Getenv("WORKER_MODE") == "READ_ONLY" {
		log.Println(" READ_ONLY: retry deshabilitado")
		delivery.Ack()
		return
	}

	payload.RetryCount++
	delay := int(math.Pow(2, float64(payload.RetryCount))) * 30

	w.client.PublishWithRetry(context.Background(), payload, delay)
	delivery.Ack()
}
*/
