package config

import "os"

const (
	DefaultWorkerMode   = "READ_WRITE"
	DefaultTicketSource = "ESTADIO"
	DefaultPacProvider  = "EDICOM"
)

type WorkerRuntimeConfig struct {
	Mode         string
	TicketSource string
	PacProvider  string
}

func LoadWorkerRuntimeConfig() WorkerRuntimeConfig {
	return WorkerRuntimeConfig{
		Mode:         getenv("WORKER_MODE", DefaultWorkerMode),
		TicketSource: getenv("WORKER_TICKET_SOURCE", DefaultTicketSource),
		PacProvider:  getenv("WORKER_PAC_PROVIDER", DefaultPacProvider),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
