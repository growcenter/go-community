package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var TYPE_COOL_MEETING = "coolMeeting"

type (
	CoolMeeting struct {
		ID             uuid.UUID
		CoolCode       string
		Name           string
		Description    *string
		MeetingDate    time.Time `gorm:"type:date"`
		MeetingStartAt string    `gorm:"type:time" column:"meeting_start_at"`
		MeetingEndAt   string    `gorm:"type:time" column:"meeting_end_at"`
		// NewJoiners     []MeetingNewJoiner `json:"new_joiners" gorm:"type:jsonb"`
		NewJoiners pq.StringArray `gorm:"type:text[]"`
		CreatedAt  *time.Time
		UpdatedAt  *time.Time
		DeletedAt  sql.NullTime
	}
	MeetingNewJoiner struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phoneNumber"`
	}
)

func (m *CreateMeetingResponse) ToResponse() *CreateMeetingResponse {
	return &CreateMeetingResponse{
		Type:           TYPE_COOL_MEETING,
		Id:             m.Id,
		Name:           m.Name,
		CoolCode:       m.CoolCode,
		Description:    m.Description,
		MeetingDate:    m.MeetingDate,
		MeetingStartAt: m.MeetingStartAt,
		MeetingEndAt:   m.MeetingEndAt,
		Members:        m.Members,
		NewJoiners:     m.NewJoiners,
	}
}

type (
	CreateMeetingRequest struct {
		MarkAttendanceNow bool                    `json:"markAttendanceNow"`
		CoolCode          string                  `json:"coolCode" validate:"required"`
		Name              string                  `json:"name"`
		Description       string                  `json:"description"`
		MeetingDate       string                  `json:"meetingDate" validate:"required,yyymmddFormat"`
		MeetingStartAt    string                  `json:"meetingStartAt" validate:"required,hhmmFormat"`
		MeetingEndAt      string                  `json:"meetingEndAt" validate:"required,hhmmFormat"`
		Attendance        CreateAttendanceRequest `json:"attendance" validate:"omitempty"`
	}
	CreateMeetingResponse struct {
		Type           string                              `json:"type"`
		Id             string                              `json:"attendanceId"`
		Name           string                              `json:"name"`
		CoolCode       string                              `json:"coolCode"`
		Description    string                              `json:"description"`
		MeetingDate    string                              `json:"meetingDate"`
		MeetingStartAt string                              `json:"meetingStartAt"`
		MeetingEndAt   string                              `json:"meetingEndAt"`
		Members        []CreateMemberAttendanceResponse    `json:"members"`
		NewJoiners     []CreateNewJoinerAttendanceResponse `json:"newJoiners"`
	}
)

func (m *GetCoolMeetingScheduleResponse) ToResponse() *GetCoolMeetingScheduleResponse {
	return &GetCoolMeetingScheduleResponse{
		Type:           TYPE_COOL_MEETING,
		Id:             m.Id,
		Name:           m.Name,
		CoolCode:       m.CoolCode,
		Description:    m.Description,
		MeetingDate:    m.MeetingDate,
		MeetingStartAt: m.MeetingStartAt,
		MeetingEndAt:   m.MeetingEndAt,
	}
}

type (
	GetPreviousUpcomingMeetingsParameter struct {
		Type        string `query:"type" validate:"required,oneof=upcoming previous"` // upcoming or previous, default upcoming if empt
		CoolCode    string `query:"coolCode" validate:"required"`
		MeetingDate string `query:"meetingDate" validate:"omitempty,yyyymmddNoExceedToday"`
	}
	GetManyByCoolCodeAndMeetingDateDBOutput struct {
		ID             uuid.UUID
		CoolCode       string
		Name           string
		Description    *string
		MeetingDate    time.Time `gorm:"type:date"`
		MeetingStartAt string    `gorm:"type:time" column:"meeting_start_at"`
		MeetingEndAt   string    `gorm:"type:time" column:"meeting_end_at"`
	}
	GetCoolMeetingScheduleResponse struct {
		Type           string `json:"type"`
		Id             string `json:"id"`
		Name           string `json:"name"`
		CoolCode       string `json:"coolCode"`
		Description    string `json:"description"`
		MeetingDate    string `json:"meetingDate"`
		MeetingStartAt string `json:"meetingStartAt"`
		MeetingEndAt   string `json:"meetingEndAt"`
	}
	GetPreviousCoolMeetingDBOutput struct {
		ID             uuid.UUID
		CoolCode       string
		Name           string
		Description    *string
		MeetingDate    time.Time `gorm:"type:date"`
		MeetingStartAt string    `gorm:"type:time" column:"meeting_start_at"`
		MeetingEndAt   string    `gorm:"type:time" column:"meeting_end_at"`
		AttendanceId   uuid.UUID
		CommunityId    string
		IsPresent      bool
		Remarks        *string
	}
	GetPreviousCoolMeetingResponse struct {
		Type           string `json:"type"`
		Id             string `json:"id"`
		Name           string `json:"name"`
		CoolCode       string `json:"coolCode"`
		Description    string `json:"description"`
		MeetingDate    string `json:"meetingDate"`
		MeetingStartAt string `json:"meetingStartAt"`
		MeetingEndAt   string `json:"meetingEndAt"`
		IsPresent      bool   `json:"isPresent"`
		Remarks        string `json:"remarks"`
	}
)
