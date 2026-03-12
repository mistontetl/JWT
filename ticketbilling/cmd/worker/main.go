package main

import (
	"log"
	"os"
	"portal_autofacturacion/data/queue"
	worker_service "portal_autofacturacion/domain/worker"

	"github.com/joho/godotenv"
)

func main() {
	//WORKER!!!
	log.Println("✅ Worker started. Waiting for request...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	queueClient, err := queue.NewClient()

	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %v", err)
	}
	defer queueClient.Close()

	worker := worker_service.NewInvoiceWorker(queueClient, os.Getenv("TICKET"), os.Getenv("PAC"))

	if err := worker.StartConsuming(queueClient); err != nil {
		log.Fatalf("Critical error when starting the Worker: %v", err)
	}

}
