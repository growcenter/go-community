package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

var (
	TYPE_COOL_ATTENDANCE = "coolAttendance"
)

type CoolAttendance struct {
	ID            uuid.UUID    `json:"id"`
	CoolMeetingId uuid.UUID    `json:"coolMeetingId"`
	CommunityId   string       `json:"communityId"`
	IsPresent     bool         `json:"isPresent"`
	Remarks       *string      `json:"remarks"`
	CreatedAt     *time.Time   `json:"createdAt"`
	UpdatedAt     *time.Time   `json:"updatedAt"`
	DeletedAt     sql.NullTime `json:"deletedAt"`
}

func (c *CreateAttendanceResponse) ToResponse() CreateAttendanceResponse {
	return CreateAttendanceResponse{
		Type:          TYPE_COOL_ATTENDANCE,
		CoolMeetingId: c.CoolMeetingId,
		Members:       c.Members,
		NewJoiners:    c.NewJoiners,
	}
}

type (
	CreateAttendanceRequest struct {
		MeetingId  uuid.UUID                          `json:"meetingId" validate:"uuid"`
		Members    []CreateMemberAttendanceRequest    `json:"members" validate:"omitempty,dive"`
		NewJoiners []CreateNewJoinerAttendanceRequest `json:"newJoiners" validate:"omitempty,dive"`
	}
	CreateMemberAttendanceRequest struct {
		CommunityId string `json:"communityId" validate:"required,communityId"`
		IsPresent   bool   `json:"isPresent"`
		Remarks     string `json:"remarks"`
	}
	CreateNewJoinerAttendanceRequest struct {
		Name        string `json:"name" validate:"required"`
		PhoneNumber string `json:"phoneNumber" validate:"required,phoneFormat0862"`
	}
	CreateAttendanceResponse struct {
		Type          string                              `json:"type"`
		CoolMeetingId string                              `json:"coolMeetingId"`
		Members       []CreateMemberAttendanceResponse    `json:"members"`
		NewJoiners    []CreateNewJoinerAttendanceResponse `json:"newJoiners"`
	}
	CreateMemberAttendanceResponse struct {
		Type         string `json:"type"`
		AttendanceId string `json:"attendanceId"`
		CommunityId  string `json:"communityId"`
		// Name         string `json:"name"`
		IsPresent bool   `json:"isPresent"`
		Remarks   string `json:"remarks"`
	}
	CreateNewJoinerAttendanceResponse struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		PhoneNumber string `json:"phoneNumber"`
	}
)
