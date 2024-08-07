package pgsql

import (
	"context"
	"fmt"
	"go-community/internal/models"
	"time"

	"gorm.io/gorm"
)

type EventRegistrationRepository interface {
	Create(ctx context.Context, eventRegistration *models.EventRegistration) (err error)
	BulkCreate(ctx context.Context, eventRegistrations *[]models.EventRegistration) (err error)
	GetAll(ctx context.Context) (eventRegistrations []models.EventRegistration, err error)
	GetAllWithParams(ctx context.Context, params models.GetAllPaginationParams) (eventRegistrations []models.GetRegisteredRepository, count int64, err error)
	GetByIdentifier(ctx context.Context, identifier string) (eventRegistrations []models.EventRegistration, err error)
	GetByCode(ctx context.Context, code string) (eventRegistration models.EventRegistration, err error)
	GetByRegisteredBy(ctx context.Context, registeredBy string) (eventRegistration []models.EventRegistration, err error)
	GetByRegisteredByStatus(ctx context.Context, registeredBy string, status string) (eventRegistration []models.EventRegistration, err error)
	GetSpecificByRegisteredBy(ctx context.Context, registeredBy string, accountNumberOrigin string) (eventRegistrations []models.GetRegisteredRepository, err error)
	BulkUpdate(ctx context.Context, eventRegistration models.EventRegistration) (err error)
	Update(ctx context.Context, eventRegistration models.EventRegistration) (err error)
	Delete(ctx context.Context, eventRegistration models.EventRegistration) (err error)
	CountSessionRegistered(ctx context.Context, sessionId string, status string) (count int64, err error)
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

func (rer *eventRegistrationRepository) GetAllWithParams(ctx context.Context, params models.GetAllPaginationParams) (eventRegistrations []models.GetRegisteredRepository, count int64, err error) {
	var results []models.GetRegisteredRepository
	var rawResults []models.GetRegisteredRaw
	var totalCount int64

	// Build the base query
	query := `
        SELECT er.*, 
               eg."name" AS general_name, 
               es."name" AS session_name
        FROM event_registrations er
        JOIN event_generals eg ON er.event_code = eg.code
        JOIN event_sessions es ON er.session_code = es.code
        WHERE 1=1
    `

	// Add filters
	if params.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", params.Search)
		query += fmt.Sprintf(" AND (er.name ILIKE '%s' OR er.registered_by ILIKE '%s' OR er.identifier ILIKE '%s')", searchPattern, searchPattern, searchPattern)
	}
	if params.FilterSessionCode != "" {
		query += fmt.Sprintf(" AND er.session_code = '%s'", params.FilterSessionCode)
	}
	if params.FilterEventCode != "" {
		query += fmt.Sprintf(" AND er.event_code = '%s'", params.FilterEventCode)
	}

	// Add sorting
	if params.Sort != "" {
		query += fmt.Sprintf(" ORDER BY %s", params.Sort)
	} else {
		query += " ORDER BY er.id DESC" // default sort
	}

	// Pagination
	offset := (params.Page - 1) * params.Limit
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", params.Limit, offset)

	// Fetch total count for pagination
	countQuery := `
        SELECT COUNT(*)
        FROM event_registrations er
        JOIN event_generals eg ON er.event_code = eg.code
        JOIN event_sessions es ON er.session_code = es.code
        WHERE 1=1
    `

	// Add the same filters to count query
	if params.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", params.Search)
		countQuery += fmt.Sprintf(" AND (er.name ILIKE '%s' OR er.registered_by ILIKE '%s' OR er.identifier ILIKE '%s')", searchPattern, searchPattern, searchPattern)
	}
	if params.FilterSessionCode != "" {
		countQuery += fmt.Sprintf(" AND er.session_code = '%s'", params.FilterSessionCode)
	}
	if params.FilterEventCode != "" {
		countQuery += fmt.Sprintf(" AND er.event_code = '%s'", params.FilterEventCode)
	}

	// Execute count query
	err = rer.db.Raw(countQuery).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	err = rer.db.Raw(query).Scan(&rawResults).Error
	if err != nil {
		return nil, 0, err
	}

	// Map raw results to composite struct
	for _, raw := range rawResults {
		result := models.GetRegisteredRepository{
			EventRegistration: models.EventRegistration{
				Name:          raw.Name,
				Identifier:    raw.Identifier,
				Address:       raw.Address,
				AccountNumber: raw.AccountNumber,
				Code:          raw.Code,
				EventCode:     raw.EventCode,
				SessionCode:   raw.SessionCode,
				RegisteredBy:  raw.RegisteredBy,
				UpdatedBy:     raw.UpdatedBy,
				Status:        raw.Status,
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

	fmt.Println(results)

	return results, totalCount, nil

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

func (rer *eventRegistrationRepository) GetByRegisteredByStatus(ctx context.Context, registeredBy string, status string) (eventRegistration []models.EventRegistration, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var ers []models.EventRegistration
	err = rer.db.Where("registered_by = ? AND status = ?", registeredBy, status).Find(&ers).Error

	return ers, err
}

func (rer *eventRegistrationRepository) GetSpecificByRegisteredBy(ctx context.Context, registeredBy string, accountNumberOrigin string) (eventRegistrations []models.GetRegisteredRepository, err error) {
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
        WHERE er.registered_by = ? AND er.account_number_origin = ?;
    `

	err = rer.db.Raw(query, registeredBy, accountNumberOrigin).Scan(&rawResults).Error
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

func (rer *eventRegistrationRepository) Delete(ctx context.Context, eventRegistration models.EventRegistration) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return rer.trx.Transaction(func(dtx *gorm.DB) error {
		return rer.db.Model(eventRegistration).Update("deleted_at", time.Now()).Error
	})
}

func (rer *eventRegistrationRepository) CountSessionRegistered(ctx context.Context, sessionId string, status string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var er models.EventRegistration
	err = rer.db.Model(er).Where("session_code = ? AND status = ?", sessionId, status).Count(&count).Error

	return count, err
}
