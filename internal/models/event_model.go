package models

import (
	"database/sql"
	"go-community/internal/common"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	TYPE_EVENT = "event"
)

type Event struct {
	ID                    int
	Code                  string
	Title                 string
	Topics                pq.StringArray `gorm:"type:text[]"`
	Description           string
	TermsAndConditions    string
	ImageLinks            pq.StringArray `gorm:"type:text[]"`
	RedirectLink          string
	CreatedBy             string
	LocationType          string
	LocationOfflineVenue  string
	LocationOnlineLink    string
	Visibility            string
	AllowedCommunityIds   pq.StringArray `gorm:"type:text[]"`
	AllowedUserTypes      pq.StringArray `gorm:"type:text[]"`
	AllowedRoles          pq.StringArray `gorm:"type:text[]"`
	AllowedCampuses       pq.StringArray `gorm:"type:text[]"`
	OrganizerCommunityIds pq.StringArray `gorm:"type:text[]"`
	Recurrence            string
	StartAt               time.Time `gorm:"type:timestamptz;not null"`
	EndAt                 time.Time `gorm:"type:timestamptz;not null"`
	PostDetails           JSONB     `gorm:"type:jsonb;default:'{}'"`
	Status                string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             sql.NullTime
}

type (
	CreateEventRequest struct {
		Title              string                       `json:"name" validate:"required" example:"Homebase"`
		Topics             []string                     `json:"topics"`
		Description        string                       `json:"description" example:"This event blabla"`
		TermsAndConditions string                       `json:"termsAndConditions" example:"This event blabla"`
		ImageLinks         []string                     `json:"imageLinks" validate:"omitempty,url"`
		RedirectLink       string                       `json:"redirectLink"`
		IsPublish          bool                         `json:"isPublish" validate:"required"`
		Location           EventLocationRequest         `json:"location" validate:"required,dive"`
		AccessConfig       EventAccessConfigRequest     `json:"accessConfig" validate:"required,dive"`
		TimeConfig         EventTimeConfigRequest       `json:"timeConfig" validate:"required,dive"`
		Questions          []BulkCreateFormQuestionItem `json:"questions" validate:"omitempty,dive"`
		Instances          []CreateInstanceRequest      `json:"instances" validate:"dive,required"`
	}
	EventLocationRequest struct {
		Type         string `json:"type" validate:"required,oneof=offline online" example:"offline"`
		OfflineVenue string `json:"offlineVenue" validate:"omitempty,required_without=OnlineLink" example:"PIOT 6 Lt. 6"`
		OnlineLink   string `json:"onlineLink" validate:"required_without=OfflineVenue,url" example:"https://www.youtube.com/watch?v=1234567890"`
	}
	EventAccessConfigRequest struct {
		Visibility   string   `json:"visibility" validate:"required,oneof=public private" example:"public"`
		CommunityIds []string `json:"communityIds" validate:"omitempty,dive,communityId" example:"community-1"`
		UserTypes    []string `json:"userTypes" validate:"omitempty" example:"volunteer"`
		Roles        []string `json:"roles" validate:"omitempty" example:"event-view-volunteer, event-edit-volunteer"`
		Campuses     []string `json:"campuses" validate:"omitempty,dive,min=3" example:"BKS"`
	}
	EventTimeConfigRequest struct {
		Recurrence string `json:"recurrence" example:"monthly"`
		StartAt    string `json:"startAt" validate:"required" example:"2024-12-10T09:02:42Z"`
		EndAt      string `json:"endAt" validate:"required" example:"2024-12-10T09:02:42Z"`
	}
	CreateEventResponse struct {
		Type               string                    `json:"type" example:"event"`
		Code               string                    `json:"code" example:"bhfe382"`
		Title              string                    `json:"title" example:"Homebase"`
		Topics             []string                  `json:"topics"`
		Description        string                    `json:"description" example:"This event blabla"`
		TermsAndConditions string                    `json:"termsAndConditions" example:"This event blabla"`
		ImageLinks         []string                  `json:"imageLinks"`
		RedirectLink       string                    `json:"redirectLink"`
		AccessConfig       EventAccessConfigResponse `json:"accessConfig"`
		TimeConfig         EventTimeConfigResponse   `json:"timeConfig"`
		Location           EventLocationResponse     `json:"location"`
		Status             string                    `json:"status" example:"available"`
		Instances          []CreateInstanceResponse  `json:"instances"`
		Questions          []FormQuestionResponse    `json:"questions,omitempty"`
	}
	EventLocationResponse struct {
		Type         string `json:"type" example:"offline"`
		OfflineVenue string `json:"offlineVenue" example:"PIOT 6 Lt. 6"`
		OnlineLink   string `json:"onlineLink" example:"https://www.youtube.com/watch?v=1234567890"`
	}
	EventAccessConfigResponse struct {
		Visibility   string   `json:"visibility"  example:"public"`
		CommunityIds []string `json:"communityIds" example:"community-1"`
		UserTypes    []string `json:"userTypes" example:"volunteer"`
		Roles        []string `json:"roles" example:"event-view-volunteer, event-edit-volunteer"`
		Campuses     []string `json:"campuses" example:"BKS"`
	}
	EventTimeConfigResponse struct {
		Recurrence string `json:"recurrence" example:"monthly"`
		StartAt    string `json:"startAt" example:"2024-12-10T09:02:42Z"`
		EndAt      string `json:"endAt" example:"2024-12-10T09:02:42Z"`
	}
)

func (e *CreateEventResponse) ToResponse() *CreateEventResponse {
	return &CreateEventResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		AccessConfig:       e.AccessConfig,
		TimeConfig:         e.TimeConfig,
		Location:           e.Location,
		Status:             e.Status,
		Instances:          e.Instances,
	}
}

func (e *GetAllEventsResponse) ToResponse() GetAllEventsResponse {
	return GetAllEventsResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		AllowedFor:         e.AllowedFor,
		AllowedUsers:       e.AllowedUsers,
		AllowedRoles:       e.AllowedRoles,
		AllowedCampuses:    e.AllowedCampuses,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		LocationType:       e.LocationType,
		ImagesLinks:        e.ImagesLinks,
		AvailabilityStatus: e.AvailabilityStatus,
	}
}

type (
	GetAllEventsDBOutput struct {
		EventCode            string         `json:"event_code"`
		EventTitle           string         `json:"event_title"`
		EventTopics          pq.StringArray `gorm:"type:text[]"`
		EventLocationType    string
		EventAllowedFor      string
		EventAllowedRoles    pq.StringArray `gorm:"type:text[]"`
		EventAllowedUsers    pq.StringArray `gorm:"type:text[]"`
		EventAllowedCampuses pq.StringArray `gorm:"type:text[]"`
		EventIsRecurring     bool           `json:"event_is_recurring"`
		EventRecurrence      string         `json:"event_recurrence"`
		EventStartAt         time.Time
		EventEndAt           time.Time
		EventRegisterStartAt time.Time `json:"event_register_start_at"`
		EventRegisterEndAt   time.Time `json:"event_register_end_at"`
		InstanceTotalSeats   int
		TotalRemainingSeats  int            `json:"total_remaining_seats"`
		EventStatus          string         `json:"event_status"`
		EventImageLinks      pq.StringArray `gorm:"type:text[]"`
		InstancesData        pq.StringArray `gorm:"type:text[]"`
	}

	GetAllEventsResponse struct {
		Type                string     `json:"type" example:"Event"`
		Code                string     `json:"code" example:"2024-HOMEBASE"`
		Title               string     `json:"title" example:"Homebase"`
		Topics              []string   `json:"topics"`
		LocationType        string     `json:"locationType" example:"offline"`
		AllowedFor          string     `json:"allowedFor" example:"public"`
		AllowedUsers        []string   `json:"allowedUsers"`
		AllowedRoles        []string   `json:"allowedRoles"`
		AllowedCampuses     []string   `json:"allowedCampuses,omitempty"`
		IsRecurring         bool       `json:"isRecurring,omitempty" example:"true"`
		Recurrence          string     `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt        time.Time  `json:"eventStartAt" example:""`
		EventEndAt          time.Time  `json:"eventEndAt" example:""`
		RegisterStartAt     *time.Time `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt       *time.Time `json:"registerEndAt,omitempty" example:""`
		TotalRemainingSeats int        `json:"totalRemainingSeats,omitempty" example:"2"`
		ImagesLinks         []string   `json:"imagesLinks"`
		AvailabilityStatus  string     `json:"availabilityStatus,omitempty" example:"available"`
	}
)

func (e *GetEventByCodeResponse) ToResponse() GetEventByCodeResponse {
	return GetEventByCodeResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		AllowedFor:         e.AllowedFor,
		AllowedUsers:       e.AllowedUsers,
		AllowedRoles:       e.AllowedRoles,
		AllowedCampuses:    e.AllowedCampuses,
		IsRecurring:        e.IsRecurring,
		Recurrence:         e.Recurrence,
		EventStartAt:       e.EventStartAt,
		EventEndAt:         e.EventEndAt,
		RegisterStartAt:    e.RegisterStartAt,
		RegisterEndAt:      e.RegisterEndAt,
		LocationType:       e.LocationType,
		LocationName:       e.LocationName,
		ImageLinks:         e.ImageLinks,
		AvailabilityStatus: e.AvailabilityStatus,
		Instances:          e.Instances,
	}
}

type (
	GetEventByCodeDBOutput struct {
		EventCode               string
		EventTitle              string
		EventTopics             pq.StringArray `gorm:"type:text[]"`
		EventDescription        string
		EventTermsAndConditions string
		EventAllowedFor         string
		EventAllowedRoles       pq.StringArray `gorm:"type:text[]"`
		EventAllowedUsers       pq.StringArray `gorm:"type:text[]"`
		EventAllowedCampuses    pq.StringArray `gorm:"type:text[]"`
		EventIsRecurring        bool
		EventRecurrence         string
		EventStartAt            time.Time
		EventEndAt              time.Time
		EventRegisterStartAt    time.Time
		EventRegisterEndAt      time.Time
		EventLocationType       string
		EventLocationName       string
		EventStatus             string
		InstanceTotalSeats      int
		TotalRemainingSeats     int            `json:"total_remaining_seats"`
		EventImageLinks         pq.StringArray `gorm:"type:text[]"`
		InstanceRegisterFlow    pq.StringArray `gorm:"type:text[]"`
		InstancesData           pq.StringArray `gorm:"type:text[]"`
	}
	GetInstanceByEventCodeDBOutput struct {
		InstanceCode              string    `json:"instance_code"`
		InstanceEventCode         string    `json:"instance_event_code"`
		InstanceTitle             string    `json:"instance_title"`
		InstanceDescription       string    `json:"instance_description"`
		InstanceStartAt           time.Time `json:"instance_start_at"`
		InstanceEndAt             time.Time `json:"instance_end_at"`
		InstanceRegisterStartAt   time.Time `json:"instance_register_start_at"`
		InstanceRegisterEndAt     time.Time `json:"instance_register_end_at"`
		InstanceAllowVerifyAt     time.Time `json:"instance_allow_verify_at"`
		InstanceDisallowVerifyAt  time.Time `json:"instance_disallow_verify_at"`
		InstanceLocationType      string    `json:"instance_location"`
		InstanceLocationName      string    `json:"instance_location_name"`
		InstanceMaxPerTransaction int       `json:"instance_max_register"`
		InstanceIsOnePerAccount   bool      `json:"instance_is_one_per_account"`
		InstanceIsOnePerTicket    bool      `json:"instance_is_one_per_ticket"`
		InstanceRegisterFlow      string    `json:"instance_register_flow"`
		InstanceCheckType         string    `json:"instance_check_type"`
		InstanceTotalSeats        int       `json:"instance_total_seats"`
		InstanceBookedSeats       int       `json:"instance_booked_seats"`
		InstanceScannedSeats      int       `json:"instance_scanned_seats"`
		InstanceStatus            string    `json:"instance_status"`
		TotalRemainingSeats       int       `json:"total_remaining_seats"`
		EventAllowedFor           string    `json:"event_allowed_for"`
	}
	GetEventByCodeParameter struct {
		Code string `json:"string" validate:"required,min=2"`
	}
	GetEventByCodeResponse struct {
		Type               string                            `json:"type" example:"event"`
		Code               string                            `json:"code" example:"bhfe382"`
		Title              string                            `json:"title" example:"Homebase"`
		Topics             []string                          `json:"topics"`
		Description        string                            `json:"description" example:"This event blabla"`
		TermsAndConditions string                            `json:"termsAndConditions" example:"This event blabla"`
		AllowedFor         string                            `json:"allowedFor" example:"public"`
		AllowedUsers       []string                          `json:"allowedUsers,omitempty"`
		AllowedRoles       []string                          `json:"allowedRoles,omitempty"`
		AllowedCampuses    []string                          `json:"allowedCampuses,omitempty"`
		IsRecurring        bool                              `json:"isRecurring" example:"true"`
		Recurrence         string                            `json:"recurrence,omitempty" example:"monthly"`
		EventStartAt       time.Time                         `json:"eventStartAt,omitempty" example:""`
		EventEndAt         time.Time                         `json:"eventEndAt,omitempty" example:""`
		RegisterStartAt    time.Time                         `json:"registerStartAt,omitempty" example:""`
		RegisterEndAt      time.Time                         `json:"registerEndAt,omitempty" example:""`
		LocationType       string                            `json:"locationType" example:"offline"`
		LocationName       string                            `json:"locationName" example:"PIOT 6 Lt. 6"`
		ImageLinks         []string                          `json:"imageLinks"`
		AvailabilityStatus string                            `json:"availabilityStatus,omitempty" example:"available"`
		Instances          []GetInstancesByEventCodeResponse `json:"instances,omitempty"`
	}
	GetInstancesByEventCodeResponse struct {
		Type                string    `json:"type" example:"eventInstance"`
		Code                string    `json:"code" example:"2024-HOMEBASE"`
		Title               string    `json:"title" example:"Homebase"`
		Description         string    `json:"description" example:"Homebase"`
		InstanceStartAt     time.Time `json:"instanceStartAt" example:""`
		InstanceEndAt       time.Time `json:"instanceEndAt" example:""`
		RegisterStartAt     time.Time `json:"registerStartAt" example:""`
		RegisterEndAt       time.Time `json:"registerEndAt" example:""`
		AllowVerifyAt       time.Time `json:"allowVerifyAt" example:""`
		DisallowVerifyAt    time.Time `json:"disallowVerifyAt" example:""`
		LocationType        string    `json:"locationType" example:"offline"`
		LocationName        string    `json:"LocationName" example:"PIOT 6 Lt. 6"`
		MaxPerTransaction   int       `json:"maxPerTransaction,omitempty"`
		IsOnePerAccount     bool      `json:"isOnePerAccount"`
		IsOnePerTicket      bool      `json:"isOnePerTicket"`
		RegisterFlow        string    `json:"registerFlow"`
		CheckType           string    `json:"checkType"`
		TotalSeats          int       `json:"totalSeats" example:"0"`
		BookedSeats         int       `json:"bookedSeats" example:"0"`
		TotalRemainingSeats int       `json:"totalRemainingSeats" example:"0"`
		AvailabilityStatus  string    `json:"availabilityStatus,omitempty" example:"available"`
	}
)

type (
	GetAllRegisteredUserDBOutput struct {
		EventCode                      string
		EventTitle                     string
		EventDescription               string
		EventTermsAndConditions        string
		EventStartAt                   time.Time
		EventEndAt                     time.Time
		EventLocationType              string
		EventLocationName              string
		EventImageLinks                pq.StringArray `gorm:"type:text[]"`
		EventStatus                    string
		InstanceCode                   string
		InstanceTitle                  string
		InstanceDescription            string
		InstanceStartAt                time.Time
		InstanceEndAt                  time.Time
		InstanceLocationType           string
		InstanceLocationName           string
		InstanceStatus                 string
		RegistrationRecordID           uuid.UUID
		RegistrationRecordName         string
		RegistrationRecordIdentifier   string
		RegistrationRecordCommunityID  string
		RegistrationRecordUpdatedBy    string
		RegistrationRecordRegisteredAt time.Time
		RegistrationRecordVerifiedAt   sql.NullTime
		RegistrationRecordStatus       string
	}
	GetAllRegisteredUserParameter struct {
		Search      string `query:"search"`
		CommunityId string `json:"communityId" validate:"required,communityId"`
	}
	GetAllRegisteredUserResponse struct {
		Type               string                                  `json:"type"`
		Code               string                                  `json:"code"`
		Title              string                                  `json:"title"`
		Description        string                                  `json:"description"`
		TermsAndConditions string                                  `json:"termsAndConditions"`
		StartAt            time.Time                               `json:"startAt"`
		EndAt              time.Time                               `json:"endAt"`
		LocationType       string                                  `json:"locationType"`
		LocationName       string                                  `json:"locationName"`
		ImageLinks         []string                                `json:"imageLinks"`
		Status             string                                  `json:"status"`
		Instances          []InstancesForRegisteredRecordsResponse `json:"instances"`
	}
	InstancesForRegisteredRecordsResponse struct {
		Type            string                          `json:"type"`
		Code            string                          `json:"code"`
		Title           string                          `json:"title"`
		Description     string                          `json:"description"`
		InstanceStartAt time.Time                       `json:"instanceStartAt"`
		InstanceEndAt   time.Time                       `json:"instanceEndAt"`
		LocationType    string                          `json:"locationType"`
		LocationName    string                          `json:"locationName"`
		Status          string                          `json:"status"`
		Registrants     []UserRegisteredRecordsResponse `json:"registrants"`
	}
	UserRegisteredRecordsResponse struct {
		Type               string    `json:"type"`
		ID                 uuid.UUID `json:"id"`
		Name               string    `json:"name"`
		Identifier         string    `json:"identifier,omitempty"`
		CommunityId        string    `json:"communityId,omitempty"`
		IdentifierOrigin   string    `json:"identifierOrigin,omitempty"`
		CommunityIdOrigin  string    `json:"communityIdOrigin,omitempty"`
		UpdatedBy          string    `json:"updatedBy,omitempty"`
		RegisteredAt       time.Time `json:"registeredAt"`
		VerifiedAt         string    `json:"verifiedAt,omitempty"`
		IsPersonalQr       bool      `json:"isPersonalQr"`
		RegistrationStatus string    `json:"registrationStatus"`
	}
)

func (e GetEventTitlesDBOutput) ToResponse() GetEventTitlesResponse {
	return GetEventTitlesResponse{
		Type:  TYPE_EVENT,
		Code:  e.Code,
		Title: e.Title,
	}
}

type (
	GetEventTitlesDBOutput struct {
		Code  string
		Title string
	}
	GetEventTitlesResponse struct {
		Type  string `json:"type" example:"event"`
		Code  string `json:"code" example:"event-1"`
		Title string `json:"title" example:"Event 1"`
	}
)

func (e GetEventSummaryDBOutput) ToResponse() *GetEventSummaryResponse {
	return &GetEventSummaryResponse{
		Type:              TYPE_EVENT,
		Code:              e.EventCode,
		Title:             e.EventTitle,
		AllowedFor:        e.EventAllowedFor,
		AllowedRoles:      e.EventAllowedRoles,
		AllowedUsers:      e.EventAllowedUsers,
		AllowedCampuses:   e.EventAllowedCampuses,
		TotalBookedSeats:  e.TotalBookedSeats,
		TotalScannedSeats: e.TotalScannedSeats,
		TotalUsers:        e.TotalUsers,
		Status:            e.EventStatus,
	}
}

func (e GetInstanceSummaryDBOutput) ToResponse() GetInstanceSummaryResponse {
	return GetInstanceSummaryResponse{
		Type:                TYPE_EVENT_INSTANCE,
		EventCode:           e.InstanceEventCode,
		Code:                e.InstanceCode,
		Title:               e.InstanceTitle,
		RegisterFlow:        e.InstanceRegisterFlow,
		CheckType:           e.InstanceCheckType,
		TotalSeats:          e.InstanceTotalSeats,
		BookedSeats:         e.InstanceBookedSeats,
		ScannedSeats:        e.InstanceScannedSeats,
		TotalRemainingSeats: e.TotalRemainingSeats,
		MaxPerTransaction:   e.InstanceMaxPerTransaction,
		AttendPercentage:    e.AttendancePercentage,
		Status:              e.InstanceStatus,
	}
}

type (
	GetEventSummaryDBOutput struct {
		EventCode            string
		EventTitle           string
		EventAllowedFor      string
		EventAllowedRoles    pq.StringArray `gorm:"type:text[]"`
		EventAllowedUsers    pq.StringArray `gorm:"type:text[]"`
		EventAllowedCampuses pq.StringArray `gorm:"type:text[]"`
		TotalBookedSeats     int
		TotalScannedSeats    int
		TotalUsers           int
		EventStatus          string
	}
	GetInstanceSummaryDBOutput struct {
		InstanceCode              string  `json:"instance_code"`
		InstanceEventCode         string  `json:"instance_event_code"`
		InstanceTitle             string  `json:"instance_title"`
		InstanceRegisterFlow      string  `json:"instance_register_flow"`
		InstanceCheckType         string  `json:"instance_check_type"`
		InstanceTotalSeats        int     `json:"instance_total_seats"`
		InstanceBookedSeats       int     `json:"instance_booked_seats"`
		InstanceScannedSeats      int     `json:"instance_scanned_seats"`
		InstanceMaxPerTransaction int     `json:"instance_max_per_transaction"`
		InstanceStatus            string  `json:"instance_status"`
		TotalRemainingSeats       int     `json:"total_remaining_seats"`
		AttendancePercentage      float64 `json:"attendance_percentage"`
	}
	GetEventSummaryResponse struct {
		Type              string   `json:"type" example:"event"`
		Code              string   `json:"code" example:"event-1"`
		Title             string   `json:"title" example:"Event 1"`
		AllowedFor        string   `json:"allowedFor" example:"volunteer"`
		AllowedRoles      []string `json:"allowedRoles" example:"event-view-volunteer, event-edit-volunteer"`
		AllowedUsers      []string `json:"allowedUsers" example:"user-1, user-2"`
		AllowedCampuses   []string `json:"allowedCampuses" example:"BKS, BKT"`
		TotalBookedSeats  int      `json:"totalBookedSeats" example:"3003"`
		TotalScannedSeats int      `json:"totalScannedSeats" example:"309"`
		TotalUsers        int      `json:"totalUsers" example:"309"`
		Status            string   `json:"status" example:"active"`
	}
	GetInstanceSummaryResponse struct {
		Type                string  `json:"type" example:"instance"`
		EventCode           string  `json:"eventCode" example:"event-1"`
		Code                string  `json:"code" example:"instance-1"`
		Title               string  `json:"title" example:"Instance 1"`
		RegisterFlow        string  `json:"registerFlow" example:"online"`
		CheckType           string  `json:"checkType" example:"online"`
		TotalSeats          int     `json:"totalSeats" example:"100"`
		BookedSeats         int     `json:"bookedSeats" example:"50"`
		ScannedSeats        int     `json:"scannedSeats" example:"50"`
		MaxPerTransaction   int     `json:"maxPerTransaction" example:"5"`
		TotalRemainingSeats int     `json:"totalRemainingSeats" example:"50"`
		AttendPercentage    float64 `json:"attendPercentage" example:"50.0"`
		Status              string  `json:"status" example:"active"`
	}
)

type EventAvailabilityStatus int32

const (
	AVAILABILITY_STATUS_AVAILABLE EventAvailabilityStatus = iota
	AVAILABILITY_STATUS_UNAVAILABLE
	AVAILABILITY_STATUS_FULL
	AVAILABILITY_STATUS_SOON
	AVAILABILITY_STATUS_WALKIN
)

const (
	AvailibilityStatusAvailable   = "available"
	AvailibilityStatusUnavailable = "unavailable"
	AvailibilityStatusFull        = "full"
	AvailibilityStatusSoon        = "soon"
	AvailibilityStatusWalkin      = "walkin"
)

var (
	MapAvailabilityStatus = map[EventAvailabilityStatus]string{
		AVAILABILITY_STATUS_AVAILABLE:   AvailibilityStatusAvailable,
		AVAILABILITY_STATUS_UNAVAILABLE: AvailibilityStatusUnavailable,
		AVAILABILITY_STATUS_FULL:        AvailibilityStatusFull,
		AVAILABILITY_STATUS_SOON:        AvailibilityStatusSoon,
		AVAILABILITY_STATUS_WALKIN:      AvailibilityStatusWalkin,
	}
)

// func DefineAvailabilityStatus(event interface{}) (string, error) {
// 	var totalRemainingSeats int
// 	var countInstanceRegisterFlows int
// 	var totalSeats int
// 	//var eventAllowedFor string
// 	var eventRegisterStartAt, eventRegisterEndAt time.Time
// 	var instanceRegisterFlows []string

// 	// Type assertion to extract fields from the concrete type
// 	switch e := event.(type) {
// 	case GetAllEventsDBOutput:
// 		totalRemainingSeats = e.TotalRemainingSeats
// 		totalSeats = e.InstanceTotalSeats
// 		instanceRegisterFlows = GetRegisterFlowsFromStringArray(e.InstancesData)
// 		countInstanceRegisterFlows = CountTotalRegisterFlows(instanceRegisterFlows)
// 		//eventAllowedFor = e.EventAllowedFor
// 		eventRegisterStartAt = e.EventRegisterStartAt
// 		eventRegisterEndAt = e.EventRegisterEndAt
// 	case *GetEventByCodeDBOutput:
// 		totalRemainingSeats = e.TotalRemainingSeats
// 		totalSeats = e.InstanceTotalSeats
// 		instanceRegisterFlows = GetRegisterFlowsFromStringArray(e.InstancesData)
// 		countInstanceRegisterFlows = CountTotalRegisterFlows(instanceRegisterFlows)
// 		//eventAllowedFor = e.EventAllowedFor
// 		eventRegisterStartAt = e.EventRegisterStartAt
// 		eventRegisterEndAt = e.EventRegisterEndAt
// 	case GetInstanceByEventCodeDBOutput:
// 		totalRemainingSeats = e.TotalRemainingSeats
// 		totalSeats = e.InstanceTotalSeats
// 		countInstanceRegisterFlows = RegisterFlowToCount(e.InstanceRegisterFlow)
// 		eventRegisterStartAt = e.InstanceRegisterStartAt
// 		eventRegisterEndAt = e.InstanceRegisterEndAt
// 		//eventAllowedFor = e.EventAllowedFor
// 		instanceRegisterFlows = []string{e.InstanceRegisterFlow}
// 	case *GetInstanceByCodeDBOutput:
// 		totalRemainingSeats = e.TotalRemainingSeats
// 		totalSeats = e.InstanceTotalSeats
// 		countInstanceRegisterFlows = RegisterFlowToCount(e.InstanceRegisterFlow)
// 		eventRegisterStartAt = e.InstanceRegisterStartAt
// 		eventRegisterEndAt = e.InstanceRegisterEndAt
// 		//eventAllowedFor = "none"
// 		instanceRegisterFlows = []string{e.InstanceRegisterFlow}
// 	default:
// 		// Return a default or error if the type is not recognized
// 		return "", ErrorInvalidInput
// 	}

// 	switch {
// 	case totalSeats == 0 && countInstanceRegisterFlows == 0:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_WALKIN], nil
// 	case totalRemainingSeats <= 0 && countInstanceRegisterFlows < len(instanceRegisterFlows):
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
// 	//case totalRemainingSeats <= 0 && countInstanceRegisterFlows == len(instanceRegisterFlows) && eventAllowedFor != "private" && totalSeats > 0:
// 	//	return MapAvailabilityStatus[AVAILABILITY_STATUS_FULL], nil
// 	case totalRemainingSeats <= 0 && countInstanceRegisterFlows == len(instanceRegisterFlows) && totalSeats > 0:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_FULL], nil
// 	case common.Now().Before(eventRegisterStartAt.In(common.GetLocation())):
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_SOON], nil
// 	case common.Now().After(eventRegisterEndAt.In(common.GetLocation())):
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_UNAVAILABLE], nil
// 	default:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
// 	}
// }

func DefineAvailabilityStatus(event interface{}) (string, error) {
	// Define a struct to hold the extracted fields
	type eventFields struct {
		totalRemainingSeats        int
		totalSeats                 int
		instanceRegisterFlows      []string
		countInstanceRegisterFlows int
		eventRegisterStartAt       time.Time
		eventRegisterEndAt         time.Time
	}

	// Extract fields based on event type
	var fields eventFields

	switch e := event.(type) {
	case GetAllEventsDBOutput:
		fields = eventFields{
			totalRemainingSeats:   e.TotalRemainingSeats,
			totalSeats:            e.InstanceTotalSeats,
			instanceRegisterFlows: GetRegisterFlowsFromStringArray(e.InstancesData),
			eventRegisterStartAt:  e.EventRegisterStartAt,
			eventRegisterEndAt:    e.EventRegisterEndAt,
		}
		fields.countInstanceRegisterFlows = CountTotalRegisterFlows(fields.instanceRegisterFlows)

	case *GetEventByCodeDBOutput:
		fields = eventFields{
			totalRemainingSeats:   e.TotalRemainingSeats,
			totalSeats:            e.InstanceTotalSeats,
			instanceRegisterFlows: GetRegisterFlowsFromStringArray(e.InstancesData),
			eventRegisterStartAt:  e.EventRegisterStartAt,
			eventRegisterEndAt:    e.EventRegisterEndAt,
		}
		fields.countInstanceRegisterFlows = CountTotalRegisterFlows(fields.instanceRegisterFlows)

	case GetInstanceByEventCodeDBOutput:
		fields = eventFields{
			totalRemainingSeats:        e.TotalRemainingSeats,
			totalSeats:                 e.InstanceTotalSeats,
			instanceRegisterFlows:      []string{e.InstanceRegisterFlow},
			countInstanceRegisterFlows: RegisterFlowToCount(e.InstanceRegisterFlow),
			eventRegisterStartAt:       e.InstanceRegisterStartAt,
			eventRegisterEndAt:         e.InstanceRegisterEndAt,
		}

	case *GetInstanceByCodeDBOutput:
		fields = eventFields{
			totalRemainingSeats:        e.TotalRemainingSeats,
			totalSeats:                 e.InstanceTotalSeats,
			instanceRegisterFlows:      []string{e.InstanceRegisterFlow},
			countInstanceRegisterFlows: RegisterFlowToCount(e.InstanceRegisterFlow),
			eventRegisterStartAt:       e.InstanceRegisterStartAt,
			eventRegisterEndAt:         e.InstanceRegisterEndAt,
		}

	default:
		// Return a default or error if the type is not recognized
		return "", ErrorInvalidInput
	}

	// Determine availability status based on extracted fields
	switch {
	case fields.totalSeats == 0 && fields.countInstanceRegisterFlows == 0:
		return MapAvailabilityStatus[AVAILABILITY_STATUS_WALKIN], nil

	case fields.totalRemainingSeats <= 0 && fields.countInstanceRegisterFlows < len(fields.instanceRegisterFlows):
		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil

	case fields.totalRemainingSeats <= 0 && fields.countInstanceRegisterFlows == len(fields.instanceRegisterFlows) && fields.totalSeats > 0:
		return MapAvailabilityStatus[AVAILABILITY_STATUS_FULL], nil

	case common.Now().Before(fields.eventRegisterStartAt.In(common.GetLocation())):
		return MapAvailabilityStatus[AVAILABILITY_STATUS_SOON], nil

	case common.Now().After(fields.eventRegisterEndAt.In(common.GetLocation())):
		return MapAvailabilityStatus[AVAILABILITY_STATUS_UNAVAILABLE], nil

	default:
		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
	}
}
