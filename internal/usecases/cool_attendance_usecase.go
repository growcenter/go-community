package usecases

import (
	"context"
	"encoding/json"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/repositories/pgsql"

	"github.com/google/uuid"
)

type CoolAttendanceUsecase interface {
	Create(ctx context.Context, request models.CreateAttendanceRequest) (res *models.CreateAttendanceResponse, err error)
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

	if &meeting.ID == nil {
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
