package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"portal_autofacturacion/data/queue"
	pac_service "portal_autofacturacion/domain/pac"
	ticket_service "portal_autofacturacion/domain/ticket"
	"portal_autofacturacion/models"
)

const (
	MaxConcurrentJobs = 3
	MaxRetries        = 3
)

type invoiceWorker struct {
	//service *billing.Service
	sem    chan struct{}
	client queue.RSGQueue

	TicketDataSource ticket_service.TicketDataSource
	PacDataSource    pac_service.PacDataSource
}

func NewInvoiceWorker(client queue.RSGQueue, ticketConfig string, pacConfig string) invoiceWorker {
	return invoiceWorker{
		//service: svc,
		sem:    make(chan struct{}, MaxConcurrentJobs),
		client: client,

		//Services!!

		PacDataSource: pac_service.NewPacDataSource(pacConfig),
	}
}

func (w invoiceWorker) HandleDelivery(delivery queue.Delivery) {

	// TODO Blocks if MaxConcurrentJobs are in progress
	w.sem <- struct{}{}

	// Start goroutine
	go func() {

		// Release traffic light when the goroutine ends
		defer func() {
			<-w.sem
		}()

		var payload models.Payload

		// Deserializar el cuerpo del mensaje
		if err := json.Unmarshal(delivery.Body(), &payload); err != nil {
			log.Printf("Error [Unmarshal]: Mensaje inválido. Descartando: %v", err)
			// Error de formato permanente. ACK para evitar reintentos.
			delivery.Ack()
			return
		}

		if payload.RetryCount >= MaxRetries {
			delivery.Ack()
			return
		}

		fmt.Printf("\nStart [%s] Ticket: %s", payload.UUID, payload.TicketFolio)
		queue.UUIDBase[payload.UUID.String()] = payload.UUID.String()
		defer delete(queue.UUIDBase, payload.UUID.String())

		// TODO How to Control an Error?
		if payload.TicketFolio == "REQUETS-10" {
			payload.RetryCount++

			fmt.Printf("\nGlue again [%s] Ticket: %s  %d", payload.UUID, payload.TicketFolio, payload.RetryCount)

			//delivery.Nack(true) // No use!!!

			if pubErr := w.client.Publish(context.Background(), payload); pubErr != nil {
				log.Printf("FATAL FAILURE: The retry message could not be sent. The message will be lost.: %v", pubErr)
			}

			delivery.Ack()
			return
		}

		time.Sleep(time.Second * 30) // Todo Simulate processes!!

		// 1) Obtain Customer Data

		// 2) TIMBRAR (PAC) Slow!!

		// 3) Save and notify (Emial!)

		// PAC!!!
		/*
			if err := w.service.ProcessInvoice(payload); err != nil {
				// Falla temporal (PAC/DB caído). NACK y Reencolar (true).
				delivery.Nack(true)
				return
			}
		*/

		fmt.Printf("\n Finish [%s] Ticket: %s", payload.UUID, payload.TicketFolio)
		delivery.Ack()
	}()

	return
}

func (w *invoiceWorker) StartConsuming(consumer queue.Consumer) error {
	return consumer.ConsumeAsync(w.HandleDelivery)
}
