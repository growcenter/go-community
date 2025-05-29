package pgsql

import (
	"context"
	"go-community/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CoolMeetingRepository interface {
	Create(ctx context.Context, meeting *models.CoolMeeting) (err error)
	Update(ctx context.Context, meeting *models.CoolMeeting) (err error)
	CheckDateByCoolCode(ctx context.Context, code string, date time.Time) (dataExist bool, err error)
	GetById(ctx context.Context, id uuid.UUID) (meeting models.CoolMeeting, err error)
	GetManyByCoolCodeAndMeetingDate(ctx context.Context, code string, date time.Time) (meetings []models.GetManyByCoolCodeAndMeetingDateDBOutput, err error)
	GetPreviousMeetings(ctx context.Context, communityId string, coolCode string, startDate time.Time, endDate time.Time) (meetings []models.GetPreviousCoolMeetingDBOutput, err error)
}

type coolMeetingRepository struct {
	db *gorm.DB
}

func NewCoolMeetingRepository(db *gorm.DB) CoolMeetingRepository {
	return &coolMeetingRepository{db: db}
}

func (cmr *coolMeetingRepository) Create(ctx context.Context, meeting *models.CoolMeeting) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return cmr.db.Create(&meeting).Error
}

func (cmr *coolMeetingRepository) Update(ctx context.Context, meeting *models.CoolMeeting) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return cmr.db.Model(&models.CoolMeeting{}).Where("id = ?", meeting.ID).Updates(meeting).Error
}

func (cmr *coolMeetingRepository) CheckDateByCoolCode(ctx context.Context, code string, date time.Time) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = cmr.db.Raw(queryCheckMeetingOnDateExistByCoolCode, code, date).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (cmr *coolMeetingRepository) GetById(ctx context.Context, id uuid.UUID) (meeting models.CoolMeeting, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()
	err = cmr.db.Find(&meeting, "id = ?", id).Error
	if err != nil {
		return meeting, err
	}
	return meeting, nil
}

func (cmr *coolMeetingRepository) GetManyByCoolCodeAndMeetingDate(ctx context.Context, code string, date time.Time) (meetings []models.GetManyByCoolCodeAndMeetingDateDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()
	err = cmr.db.Raw(queryGetMeetingsByCoolCodeAndDate, code, date).Scan(&meetings).Error
	if err != nil {
		return nil, err
	}

	return meetings, nil
}

func (cmr *coolMeetingRepository) GetPreviousMeetings(ctx context.Context, communityId string, coolCode string, startDate time.Time, endDate time.Time) (meetings []models.GetPreviousCoolMeetingDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = cmr.db.Raw(queryGetMeetingsWithAttendanceByCoolCodeAndDate, coolCode, startDate, endDate, communityId).Scan(&meetings).Error
	if err != nil {
		return nil, err
	}

	return meetings, nil
}
