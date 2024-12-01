package pgsql

import (
	"context"
	"github.com/lib/pq"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) (err error)
	GetByCode(ctx context.Context, code string) (campus models.Event, err error)
	GetAll(ctx context.Context) (campus []models.Event, err error)
	GetAllByRoles(ctx context.Context, roles []string, status string) (output []models.GetAllEventsDBOutput, err error)
	CheckByCode(ctx context.Context, code string) (dataExist bool, err error)
	GetEventAndInstancesByCode(ctx context.Context, code string) (eventWithInstances *models.GetEventByCodeDBOutput, err error)
}

type eventRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewEventRepository(db *gorm.DB, trx TransactionRepository) EventRepository {
	return &eventRepository{db: db, trx: trx}
}

func (er *eventRepository) Create(ctx context.Context, event *models.Event) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return er.trx.Transaction(func(dtx *gorm.DB) error {
		return er.db.Create(&event).Error
	})
}

func (er *eventRepository) GetByCode(ctx context.Context, code string) (campus models.Event, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e models.Event
	err = er.db.Where("code = ?", code).Find(&e).Error

	return e, err
}

func (er *eventRepository) GetAll(ctx context.Context) (campus []models.Event, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var e []models.Event
	err = er.db.Find(&e).Error

	return e, err
}

func (er *eventRepository) CheckByCode(ctx context.Context, code string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryCheckEventByCode, code).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (er *eventRepository) GetAllByRoles(ctx context.Context, roles []string, status string) (output []models.GetAllEventsDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = er.db.Raw(queryGetAllEventsByRolesAndStatus, pq.Array(roles), status).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (er *eventRepository) GetEventAndInstancesByCode(ctx context.Context, code string) (eventWithInstances *models.GetEventByCodeDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	//rows, err := er.db.Raw(queryGetEventAndInstancesByEventCode, code).Rows()
	//if err != nil {
	//	return nil, err
	//}
	//
	//defer rows.Close()
	//
	//for rows.Next() {
	//	var res models.GetEventByCodeDBOutput
	//
	//	if err := rows.Scan(
	//		&res.EventCode, &res.EventTitle, &res.EventLocation, &res.EventDescription, &res.EventCampusCode, &res.EventIsRecurring, &res.EventRecurrence, &res.EventStartAt, &res.EventEndAt, &res.EventRegisterStartAt, &res.EventRegisterEndAt, &res.EventStatus, &res.InstanceCode, &res.InstanceTitle, &res.InstanceLocation, &res.InstanceDescription, &res.InstanceStartAt, &res.InstanceEndAt, &res.InstanceRegisterStartAt, &res.InstanceRegisterEndAt, &res.InstanceMaxRegister, &res.InstanceTotalSeats, &res.InstanceBookedSeats, &res.InstanceScannedSeats, &res.InstanceStatus, &res.InstanceIsRequired, &res.TotalRemainingSeats,
	//	); err != nil {
	//		log.Println("Error scanning row:", err)
	//		return nil, err
	//	}
	//}
	//
	//return res, nil
	return nil, nil
}
