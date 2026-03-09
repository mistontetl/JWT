package repository

import (
	"context"
	"log"

	"portal_autofacturacion/models"
)

type billingRequestRepositoryReadOnly struct{}

func NewBillingRequestRepositoryReadOnly() BillingRequestRepository {
	return &billingRequestRepositoryReadOnly{}
}

func (r *billingRequestRepositoryReadOnly) Claim(token string) (bool, error) {
	log.Println(" READ_ONLY: Claim bloqueado")
	return true, nil // permite flujo sin escribir
}

func (r *billingRequestRepositoryReadOnly) Create(request *models.BillingRequest) error {
	log.Println(" READ_ONLY: Create BillingRequest bloqueado")
	return nil
}

func (r *billingRequestRepositoryReadOnly) FindByTicket(ticket string) (*models.BillingRequest, error) {
	log.Println(" READ_ONLY: FindByTicket no disponible")
	return nil, nil
}
func (r *billingRequestRepositoryReadOnly) FindByToken(token string) (*models.BillingRequest, error) {
	log.Println(" READ_ONLY: FindByToken no disponible")
	return nil, nil
}

func (r *billingRequestRepositoryReadOnly) UpdateStatus(token string, statusID int, errMsg *string) error {
	log.Println(" READ_ONLY: UpdateStatus bloqueado")
	return nil
}

func (r *billingRequestRepositoryReadOnly) ThereIsHistoryTicket(
	ctx context.Context,
	payload models.Payload,
) (models.Billing_requests, error) {
	log.Println(" READ_ONLY: ThereIsHistoryTicket bloqueado")
	return models.Billing_requests{}, nil
}

func (r *billingRequestRepositoryReadOnly) RegisterBillingHistory(
	ctx context.Context,
	payload models.Billing_requests,
) (models.Billing_requests, error) {
	log.Println(" READ_ONLY: RegisterBillingHistory bloqueado")
	return payload, nil
}
