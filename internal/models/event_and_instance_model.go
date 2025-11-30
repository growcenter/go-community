package models

import (
	"time"

	"github.com/lib/pq"
)

type GetEventAndInstanceByCodesDBOutput struct {
	EventID                    int            `gorm:"column:event_id"`
	EventCode                  string         `gorm:"column:event_code"`
	EventTitle                 string         `gorm:"column:event_title"`
	EventTopics                pq.StringArray `gorm:"column:event_topics;type:text[]"`
	EventCategory              string         `gorm:"column:event_category"`
	EventRedirectURL           string         `gorm:"column:event_redirect_url"`
	EventDescription           string         `gorm:"column:event_description"`
	EventTermsAndConditions    string         `gorm:"column:event_terms_and_conditions"`
	EventImageLinks            pq.StringArray `gorm:"column:event_image_links;type:text[]"`
	EventSlug                  string         `gorm:"column:event_slug"`
	EventCreatedBy             string         `gorm:"column:event_created_by"`
	EventLocationType          string         `gorm:"column:event_location_type"`
	EventLocationOfflineVenue  string         `gorm:"column:event_location_offline_venue"`
	EventLocationOnlineLink    string         `gorm:"column:event_location_online_link"`
	EventVisibility            string         `gorm:"column:event_visibility"`
	EventAllowedCommunityIds   pq.StringArray `gorm:"column:event_allowed_community_ids;type:text[]"`
	EventAllowedUserTypes      pq.StringArray `gorm:"column:event_allowed_user_types;type:text[]"`
	EventAllowedRoles          pq.StringArray `gorm:"column:event_allowed_roles;type:text[]"`
	EventAllowedCampuses       pq.StringArray `gorm:"column:event_allowed_campuses;type:text[]"`
	EventOrganizerCommunityIds pq.StringArray `gorm:"column:event_organizer_community_ids;type:text[]"`
	EventIsRecurring           bool           `gorm:"column:event_is_recurring"`
	EventContactCommunityIds   pq.StringArray `gorm:"column:event_contact_community_ids;type:text[]"`
	EventStartAt               time.Time      `gorm:"column:event_start_at"`
	EventEndAt                 time.Time      `gorm:"column:event_end_at"`
	EventPostDetails           JSONB          `gorm:"column:event_post_details;type:jsonb"`
	EventStatus                string         `gorm:"column:event_status"`

	InstanceID                      int            `gorm:"column:instance_id"`
	InstanceCode                    string         `gorm:"column:instance_code"`
	InstanceTitle                   string         `gorm:"column:instance_title"`
	InstanceDescription             string         `gorm:"column:instance_description"`
	InstanceParentIdentifierFields  JSONB          `gorm:"column:parent_identifier_fields;type:jsonb"`
	InstanceChildIdentifierFields   JSONB          `gorm:"column:child_identifier_fields;type:jsonb"`
	InstanceEnforceCommunityId      bool           `gorm:"column:instance_enforce_community_id"`
	InstanceEnforceSelfRegistration bool           `gorm:"column:instance_enforce_self_registration"`
	InstanceEnforceUniqueness       bool           `gorm:"column:instance_enforce_uniqueness"`
	InstanceMethods                 pq.StringArray `gorm:"column:instance_methods;type:text[]"`
	InstanceFlow                    string         `gorm:"column:instance_flow"`
	InstanceStartAt                 time.Time      `gorm:"column:instance_start_at"`
	InstanceEndAt                   time.Time      `gorm:"column:instance_end_at"`
	InstanceRegisterStartAt         time.Time      `gorm:"column:instance_register_start_at"`
	InstanceRegisterEndAt           time.Time      `gorm:"column:instance_register_end_at"`
	InstanceVerifyStartAt           time.Time      `gorm:"column:instance_verify_start_at"`
	InstanceVerifyEndAt             time.Time      `gorm:"column:instance_verify_end_at"`
	InstanceTimezone                string         `gorm:"column:instance_timezone"`
	InstanceLocationType            string         `gorm:"column:instance_location_type"`
	InstanceLocationOfflineVenue    string         `gorm:"column:instance_location_offline_venue"`
	InstanceLocationOnlineLink      string         `gorm:"column:instance_location_online_link"`
	InstanceQuotaPerUser            int            `gorm:"column:instance_quota_per_user"`
	InstanceCapacity                int            `gorm:"column:instance_capacity"`
	InstancePostDetails             JSONB          `gorm:"column:instance_post_details;type:jsonb"`
	InstanceStatus                  string         `gorm:"column:instance_status"`
}
