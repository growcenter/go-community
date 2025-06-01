package usecases

import (
	"context"
	"encoding/json"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/repositories/pgsql"
	"time"

	"github.com/google/uuid"
)

type CoolAttendanceUsecase interface {
	Create(ctx context.Context, request models.CreateAttendanceRequest) (res *models.CreateAttendanceResponse, err error)
	GetByMeetingId(ctx context.Context, meetingId uuid.UUID) (res *models.GetAllAttendanceByMeetingIdResponse, err error)
	GetSummaryByCoolCode(ctx context.Context, request models.GetSummaryAttendanceByCoolCodeRequest) (response []models.GetSummaryAttendanceByCoolCodeResponse, err error)
}

type coolAttendanceUsecase struct {
	r pgsql.PostgreRepositories
}

func NewCoolAttendanceUsecase(r pgsql.PostgreRepositories) *coolAttendanceUsecase {
	return &coolAttendanceUsecase{
		r: r,
	}
}

func (cau *coolAttendanceUsecase) Create(ctx context.Context, request models.CreateAttendanceRequest) (res *models.CreateAttendanceResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if request.MeetingId == uuid.Nil {
		return nil, models.ErrorCannotBeEmpty
	}

	meeting, err := cau.r.CoolMeeting.GetById(ctx, request.MeetingId)
	if err != nil {
		return nil, err
	}

	if meeting.ID == uuid.Nil {
		return nil, models.ErrorDataNotFound
	}

	attendanceExist, err := cau.r.CoolAttendance.CheckByMeetingId(ctx, request.MeetingId)
	if err != nil {
		return nil, err
	}

	if attendanceExist {
		return nil, models.ErrorAlreadyExist
	}

	members := make([]models.CoolAttendance, 0)
	for _, memberRequest := range request.Members {
		members = append(members, models.CoolAttendance{
			ID:            uuid.Must(uuid.NewV7()),
			CoolMeetingId: request.MeetingId,
			CommunityId:   memberRequest.CommunityId,
			IsPresent:     memberRequest.IsPresent,
			Remarks:       &memberRequest.Remarks,
		})
	}

	err = cau.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		err = cau.r.CoolAttendance.BulkCreate(ctx, &members)
		if err != nil {
			return err
		}

		if request.NewJoiners != nil {
			// newJoiners := make([]models.MeetingNewJoiner, 0)
			newJoiners := meeting.NewJoiners
			for _, newJoiner := range request.NewJoiners {
				phoneNumber, err := validator.PhoneNumber("ID", newJoiner.PhoneNumber)
				if err != nil {
					return err
				}

				jsonNewJoiner, err := json.Marshal(models.MeetingNewJoiner{
					Name:        newJoiner.Name,
					PhoneNumber: *phoneNumber,
				})
				if err != nil {
					return err
				}

				newJoiners = append(newJoiners, string(jsonNewJoiner))
			}
			meeting.NewJoiners = newJoiners

			err = cau.r.CoolMeeting.Update(ctx, &meeting)
			if err != nil {
				return err
			}
		}
		return nil
	})

	memberRes := make([]models.CreateMemberAttendanceResponse, 0)
	for _, member := range members {
		memberRes = append(memberRes, models.CreateMemberAttendanceResponse{
			Type:         models.TYPE_COOL_MEMBER,
			AttendanceId: member.ID.String(),
			CommunityId:  member.CommunityId,
			IsPresent:    member.IsPresent,
			Remarks:      *member.Remarks,
		})
	}

	// newJoinRes := make([]models.CreateNewJoinerAttendanceResponse, 0)
	// for _, newJoiner := range meeting.NewJoiners {
	// 	newJoinRes = append(newJoinRes, models.CreateNewJoinerAttendanceResponse{
	// 		Type:        models.TYPE_COOL_NEW_JOINER,
	// 		Name:        newJoiner,
	// 		PhoneNumber: newJoiner.PhoneNumber,
	// 	})
	// }

	var newJoiners []models.MeetingNewJoiner
	for _, jsonString := range meeting.NewJoiners {
		var newJoiner models.MeetingNewJoiner
		err := json.Unmarshal([]byte(jsonString), &newJoiner)
		if err != nil {
			return nil, err
		}
		newJoiners = append(newJoiners, newJoiner)
	}

	newJoinRes := make([]models.CreateNewJoinerAttendanceResponse, 0)
	for _, newJoiner := range newJoiners {
		newJoinRes = append(newJoinRes, models.CreateNewJoinerAttendanceResponse{
			Type:        models.TYPE_COOL_NEW_JOINER,
			Name:        newJoiner.Name,
			PhoneNumber: newJoiner.PhoneNumber,
		})
	}

	return &models.CreateAttendanceResponse{
		Type:          models.TYPE_COOL_MEETING,
		CoolMeetingId: meeting.ID.String(),
		Members:       memberRes,
		NewJoiners:    newJoinRes,
	}, nil
}

func (cau *coolAttendanceUsecase) GetByMeetingId(ctx context.Context, meetingId uuid.UUID) (res *models.GetAllAttendanceByMeetingIdResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	meeting, err := cau.r.CoolMeeting.GetById(ctx, meetingId)
	if err != nil {
		return nil, err
	}

	if meeting.ID == uuid.Nil {
		return nil, models.ErrorDataNotFound
	}

	attendances, err := cau.r.CoolAttendance.GetAttendancesByMeetingId(ctx, meetingId)
	if err != nil {
		return nil, err
	}

	members := make([]models.GetMemberAttendanceResponse, 0)
	presentCount := 0
	absentCount := 0
	for _, attendance := range attendances {
		members = append(members, models.GetMemberAttendanceResponse{
			Type:         models.TYPE_COOL_MEMBER,
			AttendanceId: attendance.AttendanceId.String(),
			CommunityId:  attendance.CommunityId,
			Name:         attendance.Name,
			IsPresent:    attendance.IsPresent,
			Remarks:      attendance.Remarks,
		})

		if attendance.IsPresent {
			presentCount++
		} else {
			absentCount++
		}
	}

	var newJoiners []models.CreateNewJoinerAttendanceResponse
	for _, item := range meeting.NewJoiners {
		var newJoiner models.CreateNewJoinerAttendanceResponse
		if err := json.Unmarshal([]byte(item), &newJoiner); err != nil {
			return nil, err
		}

		newJoiner.Type = models.TYPE_COOL_NEW_JOINER
		newJoiners = append(newJoiners, newJoiner)
	}

	return &models.GetAllAttendanceByMeetingIdResponse{
		Type:           models.TYPE_COOL_MEETING,
		CoolMeetingId:  meeting.ID.String(),
		Name:           meeting.Name,
		CoolCode:       meeting.CoolCode,
		Description:    *meeting.Description,
		MeetingDate:    meeting.MeetingDate.Format("2006-01-02"),
		MeetingStartAt: meeting.MeetingStartAt,
		MeetingEndAt:   meeting.MeetingEndAt,
		PresentCount:   presentCount,
		AbsentCount:    absentCount,
		Members:        members,
		NewJoiners:     newJoiners,
	}, nil
}

func (cau *coolAttendanceUsecase) GetSummaryByCoolCode(ctx context.Context, request models.GetSummaryAttendanceByCoolCodeRequest) (response []models.GetSummaryAttendanceByCoolCodeResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	existCool, err := cau.r.Cool.CheckByCode(ctx, request.CoolCode)
	if err != nil {
		return nil, err
	}

	if !existCool {
		return nil, models.ErrorDataNotFound
	}

	switch {
	case request.StartDate == "" && request.EndDate != "":
		request.StartDate = time.Now().AddDate(0, -3, 0).Format("2006-01-02")
	case request.StartDate != "" && request.EndDate == "":
		request.EndDate = time.Now().Format("2006-01-02")
	case request.StartDate == "" && request.EndDate == "":
		request.StartDate = time.Now().AddDate(0, -3, 0).Format("2006-01-02")
		request.EndDate = time.Now().Format("2006-01-02")
	default:
		// No need to self-assign StartDate since it's already set
		// Default case can be empty since StartDate and EndDate are already set
	}

	meetings, err := cau.r.CoolAttendance.GetSummaryAttendanceByCoolCode(ctx, request)
	if err != nil {
		return nil, err
	}

	for _, meeting := range meetings {
		var attendancePercentage float64
		if meeting.TotalMeetingCount > 0 {
			attendancePercentage = float64(meeting.PresentCount) / float64(meeting.TotalMeetingCount) * 100
		} else {
			attendancePercentage = 0.00
		}

		response = append(response, models.GetSummaryAttendanceByCoolCodeResponse{
			Type:                 models.TYPE_USER,
			Name:                 meeting.Name,
			CommunityId:          meeting.CommunityId,
			PresentCount:         meeting.PresentCount,
			AbsentCount:          meeting.AbsentCount,
			TotalMeetingCount:    meeting.TotalMeetingCount,
			AttendancePercentage: attendancePercentage,
		})
	}

	return response, nil
}
