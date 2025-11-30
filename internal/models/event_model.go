package models

import (
	"database/sql"
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
	Category              string
	Description           string
	TermsAndConditions    string
	ImageLinks            pq.StringArray `gorm:"type:text[]"`
	Slug                  string
	RedirectURL           string
	CreatedBy             string
	LocationType          string
	LocationOfflineVenue  string
	LocationOnlineLink    string
	LocationDetail        string
	LocationVisibility    string
	Visibility            string
	AllowedCommunityIds   pq.StringArray `gorm:"type:text[]"`
	AllowedUserTypes      pq.StringArray `gorm:"type:text[]"`
	AllowedRoles          pq.StringArray `gorm:"type:text[]"`
	AllowedCampuses       pq.StringArray `gorm:"type:text[]"`
	OrganizerCommunityIds pq.StringArray `gorm:"type:text[]"`
	ContactCommunityIds   pq.StringArray `gorm:"type:text[]"`
	IsRecurring           bool
	StartAt               time.Time `gorm:"type:timestamptz;not null"`
	EndAt                 time.Time `gorm:"type:timestamptz;not null"`
	Timezone              string    `gorm:"type:text;not null"`
	PostDetails           JSONB     `gorm:"type:jsonb"`
	Status                string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             sql.NullTime
}

// CREATE EVENT
type (
	CreateEventRequest struct {
		Title                 string                       `json:"title" validate:"required" example:"Homebase"`
		Topics                []string                     `json:"topics"`
		Category              string                       `json:"category" validate:"required"`
		Description           string                       `json:"description" example:"This event blabla"`
		TermsAndConditions    string                       `json:"termsAndConditions" example:"This event blabla"`
		ImageLinks            []string                     `json:"imageLinks" validate:"omitempty,dive,url"`
		Slug                  string                       `json:"slug,omitempty" validate:"omitempty,min=6"` // The link to redirect to after registration
		RedirectURL           string                       `json:"redirectUrl,omitempty" validate:"omitempty,url"`
		IsPublish             bool                         `json:"isPublish"`
		Location              EventLocationRequest         `json:"location" validate:"required"`
		AccessConfig          EventAccessConfigRequest     `json:"accessConfig" validate:"required"`
		TimeConfig            EventTimeConfigRequest       `json:"timeConfig" validate:"required"`
		Questions             []BulkCreateFormQuestionItem `json:"questions" validate:"omitempty,dive"`
		Instances             []CreateInstanceRequest      `json:"instances" validate:"dive,required"`
		OrganizerCommunityIds []string                     `json:"organizerCommunityIds" validate:"omitempty,dive,communityId" example:"community-1"`
		ContactCommunityIds   []string                     `json:"contactCommunityIds" validate:"required,dive,communityId" example:"community-1"`
	}
	EventLocationRequest struct {
		Type         string `json:"type" validate:"required,oneof=offline online hybrid" example:"offline"`
		OfflineVenue string `json:"offlineVenue" validate:"required_if=Type offline Type hybrid"`
		OnlineLink   string `json:"onlineLink" validate:"required_if=Type online Type hybrid,omitempty,url"`
		Detail       string `json:"detail"`
		Visibility   string `json:"visibility" validate:"omitempty,oneof=before after all" example:"public"`
	}
	EventAccessConfigRequest struct {
		Visibility   string   `json:"visibility" validate:"required,oneof=public private" example:"public"`
		CommunityIds []string `json:"communityIds" validate:"omitempty,dive,communityId" example:"community-1"`
		UserTypes    []string `json:"userTypes" validate:"omitempty" example:"volunteer"`
		Roles        []string `json:"roles" validate:"omitempty" example:"event-view-volunteer, event-edit-volunteer"`
		Campuses     []string `json:"campuses" validate:"omitempty,dive,min=3" example:"BKS"`
	}
	EventTimeConfigRequest struct {
		IsRecurring bool   `json:"isRecurring" example:"true"`
		Recurrence  string `json:"recurrence,omitempty" example:"monthly"`
		StartAt     string `json:"startAt" validate:"required" example:"2024-12-10T09:02:42Z"`
		EndAt       string `json:"endAt" validate:"required" example:"2024-12-10T09:02:42Z"`
		Timezone    string `json:"timezone" validate:"required" example:"Asia/Jakarta"`
	}
	CreateEventResponse struct {
		Type               string                    `json:"type" example:"event"`
		Code               string                    `json:"code" example:"bhfe382"`
		Title              string                    `json:"title" example:"Homebase"`
		Topics             []string                  `json:"topics"`
		Category           string                    `json:"category"`
		Description        string                    `json:"description" example:"This event blabla"`
		TermsAndConditions string                    `json:"termsAndConditions" example:"This event blabla"`
		ImageLinks         []string                  `json:"imageLinks"`
		Slug               string                    `json:"slug"`
		RedirectURL        string                    `json:"redirectUrl,omitempty"`
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
		Detail       string `json:"detail"`
		Visibility   string `json:"visibility" example:"before"`
	}
	EventAccessConfigResponse struct {
		Visibility   string   `json:"visibility"  example:"public"`
		CommunityIds []string `json:"communityIds" example:"community-1"`
		UserTypes    []string `json:"userTypes" example:"volunteer"`
		Roles        []string `json:"roles" example:"event-view-volunteer, event-edit-volunteer"`
		Campuses     []string `json:"campuses" example:"BKS"`
	}
	EventTimeConfigResponse struct {
		IsRecurring bool   `json:"isRecurring" example:"true"`
		Recurrence  string `json:"recurrence,omitempty" example:"monthly"`
		StartAt     string `json:"startAt" example:"2024-12-10T09:02:42Z"`
		EndAt       string `json:"endAt" example:"2024-12-10T09:02:42Z"`
		Timezone    string `json:"timezone" example:"Asia/Jakarta"`
	}
)

func (e *CreateEventResponse) ToResponse() *CreateEventResponse {
	return &CreateEventResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Category:           e.Category,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		ImageLinks:         e.ImageLinks,
		Slug:               e.Slug,
		AccessConfig:       e.AccessConfig,
		TimeConfig:         e.TimeConfig,
		Location:           e.Location,
		Status:             e.Status,
		Instances:          e.Instances,
		Questions:          e.Questions,
	}
}

// GET BY CODE
type (
	GetEventByCodeParameter struct {
		Code string `json:"string" validate:"required,min=5"`
	}
	GetEventWithInstancesDBOutput struct {
		EventID                    string                   `json:"event_id" db:"event_id"`
		EventCode                  string                   `json:"event_code" db:"event_code"`
		EventTitle                 string                   `json:"event_title" db:"event_title"`
		EventTopics                pq.StringArray           `json:"event_topics" db:"event_topics" gorm:"type:text[]"`
		EventCategory              string                   `json:"event_category" db:"event_category"`
		EventDescription           string                   `json:"event_description" db:"event_description"`
		EventTermsAndConditions    string                   `json:"event_terms_and_conditions" db:"event_terms_and_conditions"`
		EventImageLinks            pq.StringArray           `json:"event_image_links" db:"event_image_links" gorm:"type:text[]"`
		EventSlug                  string                   `json:"event_slug" db:"event_slug"`
		EventRedirectURL           string                   `json:"event_redirect_url" db:"event_redirect_url"`
		EventCreatedBy             string                   `json:"event_created_by" db:"event_created_by"`
		EventLocationType          string                   `json:"event_location_type" db:"event_location_type"`
		EventLocationOfflineVenue  string                   `json:"event_location_offline_venue" db:"event_location_offline_venue"`
		EventLocationOnlineLink    string                   `json:"event_location_online_link" db:"event_location_online_link"`
		EventLocationDetail        string                   `json:"event_location_detail" db:"event_location_detail"`
		EventVisibility            string                   `json:"event_visibility" db:"event_visibility"`
		EventAllowedCommunityIds   pq.StringArray           `json:"event_allowed_community_ids" db:"event_allowed_community_ids" gorm:"type:text[]"`
		EventAllowedUserTypes      pq.StringArray           `json:"event_allowed_user_types" db:"event_allowed_user_types" gorm:"type:text[]"`
		EventAllowedRoles          pq.StringArray           `json:"event_allowed_roles" db:"event_allowed_roles" gorm:"type:text[]"`
		EventAllowedCampuses       pq.StringArray           `json:"event_allowed_campuses" db:"event_allowed_campuses" gorm:"type:text[]"`
		EventOrganizerCommunityIds pq.StringArray           `json:"event_organizer_community_ids" db:"event_organizer_community_ids" gorm:"type:text[]"`
		EventContactCommunityIds   pq.StringArray           `json:"event_contact_community_ids" db:"event_contact_community_ids" gorm:"type:text[]"`
		EventIsRecurring           bool                     `json:"event_is_recurring" db:"event_is_recurring"`
		EventStartAt               time.Time                `json:"event_start_at" db:"event_start_at"`
		EventEndAt                 time.Time                `json:"event_end_at" db:"event_end_at"`
		EventPostDetails           string                   `json:"event_post_details" db:"event_post_details"`
		EventStatus                string                   `json:"event_status" db:"event_status"`
		Instances                  []InstanceDetailDBOutput `json:"instances" db:"instances" gorm:"type:jsonb;serializer:json"`
	}

	InstanceDetailDBOutput struct {
		ID                     int            `json:"id"`
		Code                   string         `json:"code"`
		EventCode              string         `json:"event_code"`
		Title                  string         `json:"title"`
		Description            string         `json:"description"`
		ParentIdentifierFields JSONB          `json:"parent_identifier_fields" gorm:"type:jsonb"`
		ChildIdentifierFields  JSONB          `json:"child_identifier_fields" gorm:"type:jsonb"`
		EnforceCommunityID     bool           `json:"enforce_community_id"`
		EnforceUniqueness      bool           `json:"enforce_uniqueness"`
		Methods                pq.StringArray `json:"methods" gorm:"type:text[]"`
		Flow                   string         `json:"flow"`
		StartAt                time.Time      `json:"start_at"`
		EndAt                  time.Time      `json:"end_at"`
		RegisterStartAt        time.Time      `json:"register_start_at"`
		RegisterEndAt          time.Time      `json:"register_end_at"`
		VerifyStartAt          time.Time      `json:"verify_start_at"`
		VerifyEndAt            time.Time      `json:"verify_end_at"`
		Timezone               string         `json:"timezone"`
		LocationType           string         `json:"location_type"`
		LocationOfflineVenue   string         `json:"location_offline_venue"`
		LocationOnlineLink     string         `json:"location_online_link"`
		LocationDetail         string         `json:"location_detail"`
		QuotaPerUser           int            `json:"quota_per_user"`
		Capacity               int            `json:"capacity"`
		PostDetails            string         `json:"post_details"`
		Status                 string         `json:"status"`

		// Added outside the DB
		PendingCount  int `json:"pending_count"`
		VerifiedCount int `json:"verified_count"`
	}

	RegistrationCount struct {
		EventCode     string `json:"event_code"`
		InstanceCode  string `json:"instance_code"`
		PendingCount  int    `json:"pending_count"`
		VerifiedCount int    `json:"verified_count"`
	}

	GetEventByCodeResponse struct {
		Type               string                            `json:"type" example:"event"`
		Code               string                            `json:"code" example:"bhfe382"`
		Title              string                            `json:"title" example:"Homebase"`
		Topics             []string                          `json:"topics"`
		Category           string                            `json:"category"`
		Description        string                            `json:"description" example:"This event blabla"`
		TermsAndConditions string                            `json:"termsAndConditions" example:"This event blabla"`
		ImageLinks         []string                          `json:"imageLinks"`
		Slug               string                            `json:"slug"`
		RedirectURL        string                            `json:"redirectUrl,omitempty"`
		Contacts           []UserIdentifierResponse          `json:"contacts"`
		Organizers         []UserIdentifierResponse          `json:"organizers"`
		CreatedBy          string                            `json:"createdBy"`
		CreatedByName      string                            `json:"createdByName"`
		Location           EventLocationResponse             `json:"location"`
		AccessConfig       EventAccessConfigResponse         `json:"accessConfig"`
		TimeConfig         EventTimeConfigResponse           `json:"timeConfig"`
		Status             string                            `json:"status" example:"available"`
		Instances          []GetInstancesByEventCodeResponse `json:"instances,omitempty"`
	}
	GetInstancesByEventCodeResponse struct {
		Type               string                             `json:"type" example:"eventInstance"`
		Code               string                             `json:"code" example:"2024-HOMEBASE"`
		EventCode          string                             `json:"eventCode" example:"2024-HOMEBASE"`
		Title              string                             `json:"title" example:"Homebase"`
		Description        string                             `json:"description" example:"Homebase"`
		IdentifierConfig   InstanceIdentifierConfigResponse   `json:"identifierConfig"`
		TimeConfig         InstanceTimeConfigResponse         `json:"timeConfig"`
		Location           EventLocationResponse              `json:"location"`
		RegistrationConfig InstanceRegistrationConfigResponse `json:"registrationConfig"`
		AvailabilityStatus string                             `json:"availabilityStatus" example:"available"`
	}
)

func (e *GetEventByCodeResponse) ToResponse() GetEventByCodeResponse {
	return GetEventByCodeResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Category:           e.Category,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		ImageLinks:         e.ImageLinks,
		Slug:               e.Slug,
		CreatedBy:          e.CreatedBy,
		CreatedByName:      e.CreatedByName,
		Location:           e.Location,
		AccessConfig:       e.AccessConfig,
		TimeConfig:         e.TimeConfig,
		Status:             e.Status,
		Instances:          e.Instances,
	}
}

// GET ALL EVENTS
func (e *GetAllEventsResponse) ToResponse() GetAllEventsResponse {
	return GetAllEventsResponse{
		Type:               TYPE_EVENT,
		Code:               e.Code,
		Title:              e.Title,
		Topics:             e.Topics,
		Category:           e.Category,
		Description:        e.Description,
		TermsAndConditions: e.TermsAndConditions,
		ImageLinks:         e.ImageLinks,
		Slug:               e.Slug,
		RedirectURL:        e.RedirectURL,
		CreatedBy:          e.CreatedBy,
		CreatedByName:      e.CreatedByName,
		Location:           e.Location,
		AccessConfig:       e.AccessConfig,
		TimeConfig:         e.TimeConfig,
		Status:             e.Status,
	}
}

type (
	GetAllEventsDBOutput struct {
		Code                  string
		Title                 string
		Topics                pq.StringArray `gorm:"type:text[]"`
		Category              string
		Description           string
		TermsAndConditions    string
		ImageLinks            pq.StringArray `gorm:"type:text[]"`
		Slug                  string
		RedirectURL           string
		CreatedBy             string
		LocationType          string
		LocationOfflineVenue  string
		LocationOnlineLink    string
		LocationDetail        string
		LocationVisibility    string
		Visibility            string
		AllowedCommunityIds   pq.StringArray `gorm:"type:text[]"`
		AllowedUserTypes      pq.StringArray `gorm:"type:text[]"`
		AllowedRoles          pq.StringArray `gorm:"type:text[]"`
		AllowedCampuses       pq.StringArray `gorm:"type:text[]"`
		OrganizerCommunityIds pq.StringArray `gorm:"type:text[]"`
		ContactCommunityIds   pq.StringArray `gorm:"type:text[]"`
		IsRecurring           bool
		StartAt               time.Time `gorm:"type:timestamptz;not null"`
		EndAt                 time.Time `gorm:"type:timestamptz;not null"`
		Timezone              string    `gorm:"type:text;not null"`
		Status                string
		CreatedAt             time.Time
		UpdatedAt             time.Time
		DeletedAt             sql.NullTime
	}

	GetAllEventsResponse struct {
		Type               string                    `json:"type" example:"event"`
		Code               string                    `json:"code" example:"bhfe382"`
		Title              string                    `json:"title" example:"Homebase"`
		Topics             []string                  `json:"topics"`
		Category           string                    `json:"category"`
		Description        string                    `json:"description" example:"This event blabla"`
		TermsAndConditions string                    `json:"termsAndConditions" example:"This event blabla"`
		ImageLinks         []string                  `json:"imageLinks"`
		Slug               string                    `json:"slug"`
		RedirectURL        string                    `json:"redirectUrl,omitempty"`
		CreatedBy          string                    `json:"createdBy" example:"1234567890"`
		CreatedByName      string                    `json:"createdByName" example:"John Doe"`
		AccessConfig       EventAccessConfigResponse `json:"accessConfig"`
		TimeConfig         EventTimeConfigResponse   `json:"timeConfig"`
		Location           EventLocationResponse     `json:"location"`
		Status             string                    `json:"status" example:"available"`
	}
)

// GET EVENT QUESTION
type (
	GetEventQuestionParameter struct {
		InstanceCode string `json:"instanceCode" validate:"required,min=15,max=15" example:"xxxxxxx-yyyyyyy"`
	}
	GetEventQuestionResponse struct {
		Type               string                             `json:"type" example:"Event"`
		EventCode          string                             `json:"eventCode" example:"2024-HOMEBASE"`
		InstanceCode       string                             `json:"instanceCode" example:"2024-HOMEBASE-20240101"`
		Title              string                             `json:"title" example:"Homebase"`
		IdentifierConfig   InstanceIdentifierConfigResponse   `json:"identifierConfig"`
		TimeConfig         InstanceTimeConfigResponse         `json:"timeConfig"`
		Location           EventLocationResponse              `json:"location"`
		RegistrationConfig InstanceRegistrationConfigResponse `json:"registrationConfig"`
		ParentQuestions    []QuestionsResponse                `json:"parentQuestions"`
		ChildQuestions     []QuestionsResponse                `json:"childQuestions"`
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

type (
	GetAllEventsParams struct {
		AllowedRoles        []string   `query:"allowedRoles" validate:"omitempty,dive,required"`
		AllowedCommunityIDs []string   `query:"allowedCommunityIds" validate:"omitempty,dive,required"`
		AllowedCampuses     []string   `query:"allowedCampuses"`
		AllowedUserTypes    []string   `query:"allowedUserTypes" validate:"omitempty,dive,required"`
		Visibility          string     `query:"visibility" validate:"omitempty,oneof=public private"`
		Status              string     `query:"status" validate:"omitempty,oneof=active inactive draft"`
		Title               string     `query:"title"`
		UserTypes           []UserType `query:"-"`
	}

	GetAllEventsResponse2 struct {
		Type                string    `json:"type" example:"event"`
		Code                string    `json:"code" example:"event-1"`
		Title               string    `json:"title" example:"Event 1"`
		Topics              []string  `json:"topics" example:"topic-1, topic-2"`
		Category            string    `json:"category"`
		ImageLinks          []string  `json:"imageLinks" example:"https://example.com/image1.jpg, https://example.com/image2.jpg"`
		Slug                string    `json:"slug" example:"event-1"`
		RedirectURL         string    `json:"redirectUrl,omitempty"`
		LocationType        string    `json:"locationType" example:"offline"`
		Visibility          string    `json:"visibility" example:"public"`
		AllowedCommunityIds []string  `json:"allowedCommunityIds" example:"BKS, BKT"`
		AllowedUserTypes    []string  `json:"allowedUserTypes" example:"volunteer"`
		AllowedRoles        []string  `json:"allowedRoles" example:"event-view-volunteer, event-edit-volunteer"`
		AllowedCampuses     []string  `json:"allowedCampuses" example:"BKS, BKT"`
		IsRecurring         bool      `json:"isRecurring" example:"true"`
		StartAt             time.Time `json:"startAt" example:"2023-01-01T00:00:00Z"`
		EndAt               time.Time `json:"endAt" example:"2023-01-01T00:00:00Z"`
		Status              string    `json:"status" example:"active"`
	}
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

// func DefineAvailabilityStatus(event interface{}) (string, error) {
// 	// Define a struct to hold the extracted fields
// 	type eventFields struct {
// 		totalRemainingSeats        int
// 		totalSeats                 int
// 		instanceRegisterFlows      []string
// 		countInstanceRegisterFlows int
// 		eventRegisterStartAt       time.Time
// 		eventRegisterEndAt         time.Time
// 	}

// 	// Extract fields based on event type
// 	var fields eventFields

// 	switch e := event.(type) {
// 	case GetAllEventsDBOutput:
// 		fields = eventFields{
// 			totalRemainingSeats:   e.TotalRemainingSeats,
// 			totalSeats:            e.InstanceTotalSeats,
// 			instanceRegisterFlows: GetRegisterFlowsFromStringArray(e.InstancesData),
// 			eventRegisterStartAt:  e.EventRegisterStartAt,
// 			eventRegisterEndAt:    e.EventRegisterEndAt,
// 		}
// 		fields.countInstanceRegisterFlows = CountTotalRegisterFlows(fields.instanceRegisterFlows)

// case *GetEventWithInstancesDBOutput:
// 		fields = eventFields{
// 			totalRemainingSeats:   e.TotalRemainingSeats,
// 			totalSeats:            e.InstanceTotalSeats,
// 			instanceRegisterFlows: GetRegisterFlowsFromStringArray(e.InstancesData),
// 			eventRegisterStartAt:  e.EventRegisterStartAt,
// 			eventRegisterEndAt:    e.EventRegisterEndAt,
// 		}
// 		fields.countInstanceRegisterFlows = CountTotalRegisterFlows(fields.instanceRegisterFlows)

// case InstanceDetailDBOutput:
// 		fields = eventFields{
// 			totalRemainingSeats:        e.TotalRemainingSeats,
// 			totalSeats:                 e.InstanceTotalSeats,
// 			instanceRegisterFlows:      []string{e.InstanceRegisterFlow},
// 			countInstanceRegisterFlows: RegisterFlowToCount(e.InstanceRegisterFlow),
// 			eventRegisterStartAt:       e.InstanceRegisterStartAt,
// 			eventRegisterEndAt:         e.InstanceRegisterEndAt,
// 		}

// 	case *GetInstanceByCodeDBOutput:
// 		fields = eventFields{
// 			totalRemainingSeats:        e.TotalRemainingSeats,
// 			totalSeats:                 e.InstanceTotalSeats,
// 			instanceRegisterFlows:      []string{e.InstanceRegisterFlow},
// 			countInstanceRegisterFlows: RegisterFlowToCount(e.InstanceRegisterFlow),
// 			eventRegisterStartAt:       e.InstanceRegisterStartAt,
// 			eventRegisterEndAt:         e.InstanceRegisterEndAt,
// 		}

// 	default:
// 		// Return a default or error if the type is not recognized
// 		return "", ErrorInvalidInput
// 	}

// 	// Determine availability status based on extracted fields
// 	switch {
// 	case fields.totalSeats == 0 && fields.countInstanceRegisterFlows == 0:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_WALKIN], nil

// 	case fields.totalRemainingSeats <= 0 && fields.countInstanceRegisterFlows < len(fields.instanceRegisterFlows):
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil

// 	case fields.totalRemainingSeats <= 0 && fields.countInstanceRegisterFlows == len(fields.instanceRegisterFlows) && fields.totalSeats > 0:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_FULL], nil

// 	case common.Now().Before(fields.eventRegisterStartAt.In(common.GetLocation())):
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_SOON], nil

// 	case common.Now().After(fields.eventRegisterEndAt.In(common.GetLocation())):
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_UNAVAILABLE], nil

// 	default:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
// 	}
// }

// func InstanceStatus(instance interface{}) (string, error) {
// 	// Define a struct to hold the extracted fields
// 	type eventFields struct {
// 		totalRemainingSeats        int
// 		capacity                   int
// 		methods                    []string
// 		countInstanceRegisterFlows int
// 		eventRegisterStartAt       time.Time
// 		eventRegisterEndAt         time.Time
// 	}

// 	// Extract fields based on event type
// 	var fields eventFields

// 	switch e := instance.(type) {
// 	case *GetEventAndInstanceByCodesDBOutput:
// 		fields = eventFields{
// 			// totalRemainingSeats:  e.TotalRemainingSeats,
// 			// totalSeats:           e.InstanceTotalSeats,
// 			methods:  e.InstanceMethods,
// 			capacity: e.InstanceCapacity,
// 			// eventRegisterStartAt: e.EventRegisterStartAt,
// 			// eventRegisterEndAt:   e.EventRegisterEndAt,
// 		}
// 	default:
// 	}

// 	// Determine availability status based on extracted fields
// 	switch {
// 	case common.CheckOneDataInList(fields.methods, []string{"direct"}) && fields.capacity == 0:
// 		return MapAvailabilityStatus[AVAILABILITY_STATUS_WALKIN], nil
// 		// case fields.totalSeats == 0 && fields.countInstanceRegisterFlows == 0:
// 		// 	return MapAvailabilityStatus[AVAILABILITY_STATUS_WALKIN], nil

// 		// case fields.totalRemainingSeats <= 0 && fields.countInstanceRegisterFlows < len(fields.methods):
// 		// 	return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil

// 		// case fields.totalRemainingSeats <= 0 && fields.countInstanceRegisterFlows == len(fields.methods) && fields.totalSeats > 0:
// 		// 	return MapAvailabilityStatus[AVAILABILITY_STATUS_FULL], nil

// 		// case common.Now().Before(fields.eventRegisterStartAt.In(common.GetLocation())):
// 		// 	return MapAvailabilityStatus[AVAILABILITY_STATUS_SOON], nil

// 		// case common.Now().After(fields.eventRegisterEndAt.In(common.GetLocation())):
// 		// 	return MapAvailabilityStatus[AVAILABILITY_STATUS_UNAVAILABLE], nil

// 		// default:
// 		// 	return MapAvailabilityStatus[AVAILABILITY_STATUS_AVAILABLE], nil
// 	}
// }
// }
