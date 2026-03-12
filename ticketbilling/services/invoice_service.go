package services

import (
	"errors"

	"portal_autofacturacion/models"

	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func (s *BillingService) CreateBillingRequest(req *models.BillingRequest) error {

	var status = "PENDING"

	err := s.db.
		Where("code = ?", "PENDING").
		First(&status).Error

	if err != nil {
		return errors.New("estado PENDING no existe")
	}

	req.StatusID = 1
	return s.db.Create(req).Error
}
