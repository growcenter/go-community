package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type EventRegistrationRepository interface {
	Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error)
	BulkCreate(ctx context.Context, eventRegistrations *[]models.EventRegistration) (err error)
	GetAll(ctx context.Context) (eventRegistrations []models.EventRegistration, err error)
	GetByIdentifier(ctx context.Context, identifier string) (eventRegistrations []models.EventRegistration, err error)
	GetByCode(ctx context.Context, code string) (eventRegistration models.EventRegistration, err error)
	GetByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistration []models.EventRegistration, err error)
	GetSpecificByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistrations []models.GetRegisteredRepository, err error)
	BulkUpdate(ctx context.Context, eventRegistration models.EventRegistration) (err error)
	Update(ctx context.Context, eventRegistration models.EventRegistration) (err error)
}

type eventRegistrationRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventRegistrationRepository(db *gorm.DB, trx TransactionRepository) EventRegistrationRepository {
	return &eventRegistrationRepository{db: db, trx: trx}
}

func (rer *eventRegistrationRepository) Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Create(&eventRegistration).Error
	})
}

func (rer *eventRegistrationRepository) BulkCreate(ctx context.Context, eventRegistrations *[]models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Create(&eventRegistrations).Error
	})
}

func (rer *eventRegistrationRepository) GetAll(ctx context.Context) (eventRegistrations []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er []models.EventRegistration
	err = rer.db.Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) GetByIdentifier(ctx context.Context, identifier string) (eventRegistrations []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er []models.EventRegistration
	err = rer.db.Where("identifier = ?", identifier).Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) GetByCode(ctx context.Context, code string) (eventRegistration models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er models.EventRegistration
	err = rer.db.Where("code = ?", code).Find(&er).Error

	return er, err
}

func (rer *eventRegistrationRepository) GetByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistration []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ers []models.EventRegistration
	err = rer.db.Where("registered_by = ?", registeredBy).Find(&ers).Error

	return ers, err
}

func (rer *eventRegistrationRepository) GetSpecificByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistrations []models.GetRegisteredRepository, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var results []models.GetRegisteredRepository
	var rawResults []models.GetRegisteredRaw

	query := `
        SELECT er.*, 
               eg."name" AS general_name, 
               es."name" AS session_name
        FROM event_registrations er
        JOIN event_generals eg ON er.event_code = eg.code
        JOIN event_sessions es ON er.session_code = es.code
        WHERE er.registered_by = $1;
    `

	err = rer.db.Raw(query, registeredBy).Scan(&rawResults).Error
	if err != nil {
		return nil, err
	}

	for _, raw := range rawResults {
		result := models.GetRegisteredRepository{
			EventRegistration: models.EventRegistration{
				ID:            int(raw.ID),
				Name:          raw.Name,
				Identifier:    raw.Identifier,
				Address:       raw.Address,
				AccountNumber: raw.AccountNumber,
				Code:          raw.Code,
				EventCode:     raw.EventCode,
				Status:        raw.Status,
				SessionCode:   raw.SessionCode,
				RegisteredBy:  raw.RegisteredBy,
			},
			EventGeneral: models.EventGeneral{
				Name: raw.GeneralName,
			},
			EventSession: models.EventSession{
				Name: raw.SessionName,
			},
		}
		results = append(results, result)
	}

	return results, err
}

func (rer *eventRegistrationRepository) BulkUpdate(ctx context.Context, eventRegistration models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		registration := models.EventRegistration{}
		return rer.db.Model(&registration).Where("id = ?", eventRegistration.ID).Updates(eventRegistration).Error
	})
}

func (rer *eventRegistrationRepository) Update(ctx context.Context, eventRegistration models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Save(eventRegistration).Error
	})
}
