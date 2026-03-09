package main

import (
	"log"
	"os"

	conexion "portal_autofacturacion/conexiones"
	"portal_autofacturacion/data/queue"
	api "portal_autofacturacion/domain/api"
	"portal_autofacturacion/domain/api/controller"
	"portal_autofacturacion/domain/api/handler"
	"portal_autofacturacion/domain/api/repository"

	"github.com/joho/godotenv"
)

// go run ./cmd/server/main.go

var queueClient queue.RSGQueue

func main() {

	// 0) Config
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//  Database
	db, err := conexion.ConexionEstadio() //ConexionBD()
	if err != nil {
		log.Fatalf("FATAL: Database: %v", err)
	}

	// TODO 1) Create Create Database Connection
	estadioDB, err := conexion.ConexionEstadio()
	if err != nil {
		log.Fatalf("FATAL: Database ESTADIO: %v", err)
	}
	// 2) RabbitMQ
	queueClient, err = queue.NewClient()

	if err != nil {
		log.Fatalf("FATAL: RabbitMQ: %v", err)
	}

	defer queueClient.Close()

	// Inject!!!

	//repo := repository.NewBillingRequestRepository( /*db*/ nil) // TODO Inject Database Connection!!!
	repo := repository.NewBillingRequestRepositoryPru(db)
	svc := controller.NewBillingHistoryController(repo, queueClient) // Logic
	hdl := handler.NewTicketHandler(svc)
	hdl.SetEstadioDB(estadioDB)

	// Config!
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := api.New(port, hdl)

	log.Fatal(app.Start())
}
