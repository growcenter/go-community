package usecases

import (
	"context"
	"fmt"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CoolMeetingUsecase interface {
	Create(ctx context.Context, request models.CreateMeetingRequest) (response *models.CreateMeetingResponse, err error)
	GetUpcomingMeetings(ctx context.Context, parameter models.GetPreviousUpcomingMeetingsParameter) (response []models.GetCoolMeetingScheduleResponse, err error)
	GetPreviousMeetings(ctx context.Context, communityId string, parameter models.GetPreviousUpcomingMeetingsParameter) (response []models.GetPreviousCoolMeetingResponse, err error)
}

type coolMeetingUsecase struct {
	r   pgsql.PostgreRepositories
	cfg config.Configuration
	a   CoolAttendanceUsecase
}

func NewCoolMeetingUsecase(r pgsql.PostgreRepositories, cfg config.Configuration, a CoolAttendanceUsecase) *coolMeetingUsecase {
	return &coolMeetingUsecase{
		r:   r,
		cfg: cfg,
		a:   a,
	}
}

func (cmu *coolMeetingUsecase) Create(ctx context.Context, request models.CreateMeetingRequest) (response *models.CreateMeetingResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	cool, err := cmu.r.Cool.GetNameByCode(ctx, request.CoolCode)
	if err != nil {
		return nil, err
	}

	if cool.Code == "" {
		return nil, models.ErrorDataNotFound
	}

	meetingDate, err := time.Parse("2006-01-02", request.MeetingDate)
	if err != nil {
		return nil, err
	}

	existMeetingOnDate, err := cmu.r.CoolMeeting.CheckDateByCoolCode(ctx, request.CoolCode, meetingDate)
	if err != nil {
		return nil, err
	}

	if existMeetingOnDate {
		return nil, models.ErrorAlreadyExist
	}

	startAt, err := time.Parse("15:04", request.MeetingStartAt)
	if err != nil {
		return nil, err
	}

	endAt, err := time.Parse("15:04", request.MeetingEndAt)
	if err != nil {
		return nil, err
	}

	if startAt.After(endAt) {
		return nil, models.ErrorStartDateLater
	}

	if endAt.Before(startAt) {
		return nil, models.ErrorInvalidInput
	}

	if request.Name == "" {
		request.Name = fmt.Sprintf("COOL %s Meeting - %s", cool.Name, meetingDate.Format("02 Jan 2006"))
	}

	meeting := models.CoolMeeting{
		ID:             uuid.Must(uuid.NewV7()),
		Name:           request.Name,
		Topic:          request.Topic,
		Description:    &request.Description,
		LocationType:   request.LocationType,
		LocationName:   request.LocationName,
		CoolCode:       request.CoolCode,
		MeetingDate:    meetingDate,
		MeetingStartAt: startAt.Format("15:04"),
		MeetingEndAt:   endAt.Format("15:04"),
	}

	var attendanceRes *models.CreateAttendanceResponse
	err = cmu.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		err = cmu.r.CoolMeeting.Create(ctx, &meeting)
		if err != nil {
			return err
		}

		if request.MarkAttendanceNow {
			attendanceRequest := models.CreateAttendanceRequest{
				MeetingId:  meeting.ID,
				Members:    request.Attendance.Members,
				NewJoiners: request.Attendance.NewJoiners,
			}

			attendanceRes, err = cmu.a.Create(ctx, attendanceRequest)
			if err != nil {
				return err
			}
		}

		return nil
	})

	response = &models.CreateMeetingResponse{
		Type:           models.TYPE_COOL_MEETING,
		Id:             meeting.ID.String(),
		Name:           meeting.Name,
		Topic:          meeting.Topic,
		Description:    *meeting.Description,
		LocationType:   meeting.LocationType,
		LocationName:   meeting.LocationName,
		CoolCode:       meeting.CoolCode,
		MeetingDate:    meeting.MeetingDate.Format("2006-01-02"),
		MeetingStartAt: meeting.MeetingStartAt,
		MeetingEndAt:   meeting.MeetingEndAt,
		Members:        attendanceRes.Members,
		NewJoiners:     attendanceRes.NewJoiners,
	}

	return response, nil
}

func (cmu *coolMeetingUsecase) GetUpcomingMeetings(ctx context.Context, parameter models.GetPreviousUpcomingMeetingsParameter) (response []models.GetCoolMeetingScheduleResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if parameter.MeetingDate == "" {
		parameter.MeetingDate = time.Now().Format("2006-01-02")
	}

	meetingDate, err := time.Parse("2006-01-02", parameter.MeetingDate)
	if err != nil {
		return nil, err
	}
	meetings, err := cmu.r.CoolMeeting.GetManyByCoolCodeAndMeetingDate(ctx, parameter.CoolCode, meetingDate)
	if err != nil {
		return nil, err
	}

	for _, meeting := range meetings {
		response = append(response, models.GetCoolMeetingScheduleResponse{
			Type:           models.TYPE_COOL_MEETING,
			Id:             meeting.ID.String(),
			Name:           meeting.Name,
			Topic:          meeting.Topic,
			CoolCode:       meeting.CoolCode,
			Description:    strings.TrimSpace(*meeting.Description),
			LocationType:   meeting.LocationType,
			LocationName:   meeting.LocationName,
			MeetingDate:    meeting.MeetingDate.Format("2006-01-02"),
			MeetingStartAt: meeting.MeetingStartAt,
			MeetingEndAt:   meeting.MeetingEndAt,
		})
	}

	return response, nil
}

func (cmu *coolMeetingUsecase) GetPreviousMeetings(ctx context.Context, communityId string, parameter models.GetPreviousUpcomingMeetingsParameter) (response []models.GetPreviousCoolMeetingResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if parameter.MeetingDate == "" {
		parameter.MeetingDate = time.Now().Format("2006-01-02")
	}

	meetingEndDate, err := time.Parse("2006-01-02", parameter.MeetingDate)
	if err != nil {
		return nil, err
	}

	threeMonthsAgo := meetingEndDate.AddDate(0, -cmu.cfg.Cool.PreviousDateMeeting, 0)
	meetingStartDate := time.Date(threeMonthsAgo.Year(), threeMonthsAgo.Month(), 1, 0, 0, 0, 0, meetingEndDate.Location())
	meetings, err := cmu.r.CoolMeeting.GetPreviousMeetings(ctx, communityId, parameter.CoolCode, meetingStartDate, meetingEndDate)
	if err != nil {
		return nil, err
	}

	for _, meeting := range meetings {
		response = append(response, models.GetPreviousCoolMeetingResponse{
			Type:           models.TYPE_COOL_MEETING,
			Id:             meeting.ID.String(),
			Name:           meeting.Name,
			Topic:          meeting.Topic,
			CoolCode:       meeting.CoolCode,
			Description:    *meeting.Description,
			LocationType:   meeting.LocationType,
			LocationName:   meeting.LocationName,
			MeetingDate:    meeting.MeetingDate.Format("2006-01-02"),
			MeetingStartAt: meeting.MeetingStartAt,
			MeetingEndAt:   meeting.MeetingEndAt,
			IsPresent:      meeting.IsPresent,
			Remarks:        *meeting.Remarks,
		})
	}

	return response, nil
}
