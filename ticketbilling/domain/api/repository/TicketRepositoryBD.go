package repository

import (
	"errors"
	"time"

	"portal_autofacturacion/models"

	"gorm.io/gorm"
)

type ticketRepositoryDB struct {
	db *gorm.DB
}

func NewTicketRepositoryDB(db *gorm.DB) TicketStoreRepository {
	return &ticketRepositoryDB{db: db}
}

func (r *ticketRepositoryDB) Upsert(t models.Ticket) (models.Ticket, error) {
	var dbT models.Ticket

	err := r.db.
		Where("source_ticket_id = ? AND source_system = ?", *&t.SourceSystem).
		First(&dbT).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return t, r.db.Create(&t).Error
	}

	if err != nil {
		return models.Ticket{}, err
	}

	err = r.db.Model(&dbT).
		Updates(map[string]interface{}{
			"total_amount": t.TotalAmount,
			"updated_at":   time.Now(),
		}).Error

	return dbT, err
}
