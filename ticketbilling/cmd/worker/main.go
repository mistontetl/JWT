package main

import (
	//	"context"
	//	"fmt"
	"log"
	//	"time"

	conexion "portal_autofacturacion/conexiones"
	"portal_autofacturacion/config"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/domain/worker"

	//	"portal_autofacturacion/models"

	//	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("✅ Worker started. Waiting for request...")

	// ─────────────────────────────────────────────
	// ENV
	// ─────────────────────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := conexion.ConexionEstadio() //ConexionBD()
	if err != nil {
		log.Fatalf("Error de conexión a la DB: %v", err)
	}

	queueClient, err := queue.NewClient()
	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %v", err)
	}
	defer queueClient.Close()

	/*
		payloadPrueba := models.Payload{
			UUID:        uuid.New(),
			TicketFolio: "308460", // ← folio real de ESTADIO
			Total:       806200,
		}
	*/
	//	fmt.Printf(" Enviando mensaje para Ticket: %s...\n", payloadPrueba.TicketFolio)

	//ctxPub, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	//if err := queueClient.Publish(ctxPub, payloadPrueba); err != nil {
	//	log.Printf(" Error al publicar mensaje de prueba: %v", err)
	//} else {
	//		log.Println(" Mensaje de prueba enviado a la cola.")
	//	}

	// ─────────────────────────────────────────────
	// WORKER
	// ─────────────────────────────────────────────
	runtimeConfig := config.LoadWorkerRuntimeConfig()

	invoiceWorker := worker.NewInvoiceWorker(
		queueClient,
		runtimeConfig,
		db,
	)

	log.Println(" Worker listo, iniciando consumo...")

	if err := invoiceWorker.StartConsuming(queueClient); err != nil {
		log.Fatalf(" Error crítico al iniciar el Worker: %v", err)
	}

	log.Println(" Worker en ejecución. Esperando mensajes...")

	select {}
}

/*package main

import (
	//	"context"

	"context"
	"fmt"
	"log"
	"time"

	conexion "portal_autofacturacion/conexiones"
"portal_autofacturacion/config"
	"portal_autofacturacion/data/queue"
	"portal_autofacturacion/migrations"
	"portal_autofacturacion/models"

	"portal_autofacturacion/domain/worker"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	//WORKER!!!
	log.Println("✅ Worker started. Waiting for request...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	////

	db, err := conexion.ConexionBD()
	if err != nil {
		log.Fatalf("Error de conexión a la DB: %v", err)
	}
	///

	migrations.RunBillingMigrations(db)

	queueClient, err := queue.NewClient()

	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %v", err)
	}
	defer queueClient.Close()

	tokenStr := "0d8f447b-e99e-46b1-9e7b-34eb203c5ebd"

	payloadPrueba := models.Payload{
		UUID:        uuid.MustParse(tokenStr),
		TicketFolio: "9001",
		Total:       382.52,
	}

	fmt.Printf("Enviando mensaje para Ticket: %s...\n", payloadPrueba.TicketFolio)
	ctxPub, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := queueClient.Publish(ctxPub, payloadPrueba); err != nil {
		log.Printf(" Error al publicar mensaje de prueba: %v", err)
	} else {
		fmt.Println(" Mensaje de prueba enviado a la cola.")
	}
	if err != nil {
		log.Fatalf(" Error al publicar: %v", err)
	}

	fmt.Println(" Mensaje publicado con éxito.")

	///
	invoiceWorker := worker.NewInvoiceWorker(queueClient, "ESTADIO", "PAC", db)
	//invoiceWorker := worker.NewInvoiceWorker(queueClient, "CCO", "PAC", db)

	if err := invoiceWorker.StartConsuming(queueClient); err != nil {
		log.Fatalf(" Error crítico al iniciar el Worker: %v", err)
	}

	log.Println(" Worker en ejecución. Esperando mensajes de RabbitMQ...")

	select {}
	///

}
*/
/*
package main

import (
	"fmt"
	"log"

	conexion "portal_autofacturacion/conexiones"
"portal_autofacturacion/config"
	ticket_service "portal_autofacturacion/domain/ticket/service"
)

func main() {

	fmt.Println("PRUEBA DIRECTA ESTADIO")

	// Conexión a PostgreSQL Azure (Estadio)
	dbEstadio, err := conexion.ConexionEstadio()
	if err != nil {
		log.Fatalf("Error conectando a Estadio: %v", err)
	}

	// DB local solo por contrato (NO se usa)
	dbLocal := dbEstadio

	// Crear adapter de Estadio
	estadio := ticket_service.NewEstadioTicket(dbEstadio, dbLocal)

	// Folio de prueba
	folio := "308460"

	fmt.Println(" Consultando ticket en ESTADIO:", folio)

	ticket, err := estadio.GetValidTicketForBilligS(folio)
	if err != nil {
		log.Fatalf(" Error consultando estadio: %v", err)
	}

	fmt.Println(" Ticket obtenido desde ESTADIO:")
	fmt.Printf("%+v\n", ticket)
}
*/
