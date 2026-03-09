package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/domain/api/repository"
	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"

	"github.com/google/uuid"
)

/*
	TODO  -----------------  LOGIC -----------------
*/
///cambio

type BillingRequestController struct {
	repo  repository.BillingRequestRepository
	queue queue.RSGQueue
}

func NewBillingHistoryController(r repository.BillingRequestRepository, q queue.RSGQueue) *BillingRequestController {
	return &BillingRequestController{repo: r, queue: q}
}

//func (s *BillingRequestController) ProcessRequest(ctx context.Context, payload models.Payload) (string, error) {

// TODO HERE!!!! Implement logic here (Points 4, 5, 6)
func (c *BillingRequestController) CreateInvoice(req models.Payload) (p models.BillingRequest, b bool, u *uuid.UUID, er error) {

	p = models.BillingRequest{}
	// ================================
	// 1. VALIDACIONES
	// ================================
	fmt.Println(" Validando ticket", req)

	cleanTicket := utils.Normalize(req.TicketFolio)
	if cleanTicket == "" {
		fmt.Println(" Error: ticket vacío")
		return p, false, nil, errors.New("el ticket es obligatorio")
	}

	if utils.HasSpecialChars(cleanTicket) {
		fmt.Println(" Error: ticket con caracteres especiales")
		return p, false, nil, errors.New("el ticket contiene caracteres especiales no permitidos")
	}

	fmt.Println(" Ticket válido:", cleanTicket)

	// -------------------------
	fmt.Println(" Validando email")
	cleanEmail := utils.Space(req.Email)
	if !utils.IsValidEmail(cleanEmail) {
		fmt.Println(" Error: email inválido")
		return p, false, nil, errors.New("el formato del correo electrónico es inválido")
	}
	fmt.Println(" Email válido:", cleanEmail)

	// -------------------------F
	fmt.Println(" Normalizando RFC")
	cleanRFC := utils.Normalize(req.RFC)
	fmt.Println(" Normalizando RFC ========== ", cleanRFC)
	if !utils.IsValidRFC(cleanRFC) {
		fmt.Println(" Error: RFC inválido", cleanRFC)
		return p, false, nil, errors.New("el formato del RFC inválido")
	}
	// -------------------------
	//fmt.Println(" Validando monto")
	if req.Total <= 0 {
		fmt.Println(" Error: monto inválido")
		return p, false, nil, errors.New("monto inválido")
	}
	//fmt.Println("", req.Total)

	// ==========================
	// 2. DEDUPLICACIÓN
	// ==========================
	fmt.Println(" Buscando ticket existente en la BD")

	existing, err := c.repo.FindByTicket(cleanTicket)
	if err != nil {
		return p, false, nil, err
	}

	if existing != nil {
		log.Printf(
			"[Create] Ticket existente encontrado ticket=%s token=%s status=%s rfc_db=%s rfc_req=%s",
			existing.UserInputTicket,
			existing.RequestToken,
			//		existing.Status.Code,
			existing.RFC,
			req.RFC,
		)

		//	id, err := uuid.Parse(existing.RequestToken)

		if err != nil {
			log.Printf(
				"[Create][ERROR] UUID inválido token=%s err=%v",
				existing.RequestToken,
				err,
			)
			return p, false, nil, errors.New("Error interno al recuperar llave")
		}
		/*
			switch existing.Status.Code {

			// ============================
			// B) SUCCESS / MAIL_FAILED
			// ============================
			case "SUCCESS", "MAIL_FAILED":
				log.Printf(
					"[Create] REUTILIZANDO token=%s status=%s",
					existing.RequestToken,
					existing.Status.Code,
				)
				return *existing, false, &id, nil

			// ============================
			// C) Estados intermedios
			// ============================
			case "PENDING", "VALIDATING", "STAMPING":
				if existing.RFC == req.RFC {
					log.Printf(
						"[Create] EN_PROCESO token=%s status=%s (RFC coincide)",
						existing.RequestToken,
						existing.Status.Code,
					)
					return *existing, false, &id, nil
				}
				log.Printf(
					"[Create][CONFLICT] RFC distinto token=%s rfc_db=%s rfc_req=%s",
					existing.RequestToken,
					existing.RFC,
					req.RFC,
				)
				return p, false, nil, fmt.Errorf("409 Conflict: RFC distinto")

			// ============================
			// D) ERROR → permitir nuevo
			// ============================
			case "ERROR":
				log.Printf(
					"[CreateInvoice] Estado ERROR previo, se permitirá crear nuevo ticket=%s",
					existing.UserInputTicket,
				)
				// continuar y crear nuevo
			} */
	}

	// ==========================
	// 3. GENERAR TOKEN
	// ==========================
	requestToken := uuid.New()

	fmt.Println(" Token generado:", requestToken)

	// ==========================
	// 4. GUARDAR EN BD (PENDING)
	// ==========================
	fmt.Println(" Guardando billing_request en estado PENDING")

	br := models.BillingRequest{
		RequestToken:    requestToken.String(),
		UserInputTicket: cleanTicket,
		TotalAmount:     req.Total,
		//	ClientEmail:     cleanEmail,
		RFC: cleanRFC,
		//	StatusID:     utils.PENDING,
		SourceSystem: "",
	}

	if err := c.repo.Create(&br); err != nil {
		fmt.Println(" Error al insertar en billing_requests:", err)
		return p, false, nil, err
	}

	if existing != nil {
		log.Printf(
			"[Create] EXISTENTE token=%s ticket=%s status=%s",
			existing.RequestToken,
			existing.UserInputTicket,
			//		existing.Status.Code,
		)
	}

	fmt.Println(" Registro creado correctamente en BD")

	// ==========================
	//  RESPUESTA
	// ==========================
	fmt.Println(" Proceso CreateInvoice finalizado")

	return br, true, &requestToken, nil
}

/*
=================================================
PUBLISH TO QUEUE
=================================================
*/
func (c *BillingRequestController) GetInvoiceStatus(token string) (*models.BillingRequest, error) {
	if token == "" {
		return nil, errors.New("el request_token es obligatorio")
	}

	return c.repo.FindByToken(token)
} ////

func (c *BillingRequestController) PublishToQueue(ctx context.Context, payload models.Payload) error {

	fmt.Println(" Publicando mensaje en la cola")

	if c.queue == nil {
		fmt.Println(" Error: cola no inicializada")
		return fmt.Errorf("la conexión a la cola no está inicializada")
	}

	err := c.queue.Publish(ctx, payload)
	if err != nil {
		fmt.Println(" Error al publicar en la cola:", err)
		return err
	}

	fmt.Println(" Mensaje enviado correctamente a la cola")
	return nil
}

func (s *BillingRequestController) ProcessRequest(ctx context.Context, payload models.Payload) (string, error) {

	fmt.Println(" Iniciando ProcessRequest (worker)")

	fmt.Println(" Registrando historial en BD con estado PENDING")
	billing, err := s.repo.RegisterBillingHistory(ctx, models.Billing_requests{
		Status: utils.PENDING,
	})

	if err != nil {
		fmt.Println(" Error al registrar historial:", err)
		return "", err
	}

	fmt.Println(" Historial registrado con ID:", billing.ID)

	// ==========================
	// Preparar payload
	// ==========================
	payload.UUID = uuid.New()
	//	payload.BillingRequestID = billing.ID

	fmt.Println(" UUID generado:", payload.UUID.String())

	// ==========================
	// Enviar a la cola
	// ==========================
	fmt.Println(" Enviando payload a la cola")
	err = s.queue.Publish(ctx, payload)

	if err != nil {
		fmt.Println(" Error al publicar en la cola:", err)
		return "", err
	}

	fmt.Println(" ProcessRequest finalizado correctamente")

	// Se retorna UUID para polling
	return payload.UUID.String(), nil
}
