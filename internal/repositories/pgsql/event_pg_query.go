package pgsql

import (
	"fmt"
	"go-community/internal/common"
	"go-community/internal/models"
	"strings"

	"github.com/lib/pq"
)

const (
	queryGetAllEvents = `
		SELECT
			e.code AS event_code,
			e.title AS event_title,
			e.topics AS event_topics,
			e.category AS event_category,
			e.description AS event_description,
			e.terms_and_conditions AS event_terms_and_conditions,
			e.slug AS event_slug,
			e.redirect_url AS event_redirect_url,
			e.created_by AS event_created_by,
			e.location_type AS event_location_type,
			e.location_offline_venue AS event_location_offline_venue,
			e.location_online_link AS event_location_online_link,
			e.location_detail AS event_location_detail,
			e.location_visibility AS event_location_visibility,
			e.visibility as event_allowed_for,
			e.allowed_roles AS event_allowed_roles,
			e.allowed_community_ids AS event_allowed_users,
			e.allowed_campuses AS event_allowed_campuses,
			e.allowed_user_types AS event_allowed_user_types,
			e.organizer_community_ids AS event_organizer_community_ids,
			e.contact_community_ids AS event_contact_community_ids,
			e.is_recurring AS event_is_recurring,
			e.recurrence AS event_recurrence,
			e.start_at AS event_start_at,
			e.end_at AS event_end_at,
			COALESCE(e.image_links, ARRAY[]::TEXT[]) AS event_image_links,
			e.status AS event_status,
		FROM
			events e
		LEFT JOIN
			event_instances i ON e.code = i.event_code
	`

	queryGetEventWithInstancesByCode = `
		SELECT
			e.id AS event_id,
			e.code AS event_code,
			e.title AS event_title,
			e.category AS event_category,
			e.topics AS event_topics,
			e.description AS event_description,
			e.terms_and_conditions AS event_terms_and_conditions,
			e.image_links AS event_image_links,
			e.slug AS event_slug,
			e.redirect_url AS event_redirect_url,
			e.created_by AS event_created_by,
			e.location_type AS event_location_type,
			e.location_offline_venue AS event_location_offline_venue,
			e.location_online_link AS event_location_online_link,
			e.location_detail AS event_location_detail,
			e.visibility AS event_visibility,
			e.allowed_community_ids AS event_allowed_community_ids,
			e.allowed_user_types AS event_allowed_user_types,
			e.allowed_roles AS event_allowed_roles,
			e.allowed_campuses AS event_allowed_campuses,
			e.organizer_community_ids AS event_organizer_community_ids,
			e.contact_community_ids AS event_contact_community_ids, e.is_recurring AS event_is_recurring,
			e.start_at AS event_start_at,
			e.end_at AS event_end_at,
			e.post_details AS event_post_details,
			e.status AS event_status,
			json_agg(
				json_build_object(
					'id', ei.id,
					'code', ei.code,
					'event_code', ei.event_code,
					'title', ei.title,
					'description', ei.description,
					'parent_identifier_fields', ei.parent_identifier_fields,
					'child_identifier_fields', ei.child_identifier_fields,
					'enforce_community_id', ei.enforce_community_id,
					'enforce_uniqueness', ei.enforce_uniqueness,
					'enforce_self_registration', ei.enforce_self_registration,
					'methods', ei.methods,
					'flow', ei.flow,
					'start_at', ei.start_at,
					'end_at', ei.end_at,
					'register_start_at', ei.register_start_at,
					'register_end_at', ei.register_end_at,
					'verify_start_at', ei.verify_start_at,
					'verify_end_at', ei.verify_end_at,
					'timezone', ei.timezone,
					'location_type', ei.location_type,
					'location_offline_venue', ei.location_offline_venue,
					'location_online_link', ei.location_online_link,
					'quota_per_user', ei.quota_per_user,
					'capacity', ei.capacity,
					'post_details', ei.post_details,
					'status', ei.status
				)
			) FILTER (WHERE ei.id IS NOT NULL) AS instances
		FROM
			events e
		LEFT JOIN
			event_instances ei ON e.code = ei.event_code
		WHERE
			e.%s = ? AND e.deleted_at IS NULL
		GROUP BY
			e.id;
	`
)

var (
	queryCheckEventByCode = "SELECT EXISTS (SELECT 1 FROM events WHERE code = ?)"

	queryGetRegisteredUserByCommunityIdOrigin = `
	SELECT DISTINCT 
		e.code AS event_code,
		e.title AS event_title,
		e.description AS event_description,
		e.terms_and_conditions AS event_terms_and_conditions,
		e.event_start_at AS event_start_at,
		e.event_end_at AS event_end_at,
		e.location_type AS event_location_type,
		e.location_name AS event_location_name,
		COALESCE(e.image_links, ARRAY[]::TEXT[]) AS event_image_links, -- Default to empty array
		e.status AS event_status,
		ei.code AS instance_code,
		ei.title AS instance_title,
		ei.description AS instance_description,
		ei.instance_start_at AS instance_start_at,
		ei.instance_end_at AS instance_end_at,
		ei.location_type AS instance_location_type,
		ei.location_name AS instance_location_name,
		ei.status AS instance_status,
		rr.id AS registration_record_id,
		rr.name AS registration_record_name,
		coalesce(rr.identifier, '') AS registration_record_identifier,
		coalesce(rr.community_id, '') AS registration_record_community_id,
		coalesce(rr.updated_by, '') AS registration_record_updated_by,
		rr.registered_at AS registration_record_registered_at,
		rr.verified_at as registration_record_verified_at,
		rr.status AS registration_record_status
	FROM
		events e
		JOIN
			event_instances ei ON e.code = ei.event_code
		JOIN
			event_registration_records rr ON rr.instance_code = ei.code
	WHERE
		rr.community_id_origin = ?
`
	queryGetEventTitles = `SELECT code, title FROM events WHERE deleted_at IS NULL`

	queryGetEventSummary = `
		SELECT
			e.code AS event_code,
			e.title AS event_title,
			e.allowed_for AS event_allowed_for, -- Non-nullable
			COALESCE(e.allowed_roles, ARRAY[]::TEXT[]) AS event_allowed_roles, -- Default to empty array
			COALESCE(e.allowed_users, ARRAY[]::TEXT[]) AS event_allowed_users, -- Default to empty array
			COALESCE(e.allowed_campuses, ARRAY[]::TEXT[]) AS event_allowed_campuses, -- Default to empty array
			COALESCE(SUM(ei.booked_seats), 0) AS total_booked_seats,
			COALESCE(SUM(ei.scanned_seats), 0) AS total_scanned_seats,
			e.status AS event_status
		FROM
			events e
		LEFT JOIN
			event_instances ei ON e.code = ei.event_code
		WHERE
			e.code = ?
		  AND e.deleted_at IS NULL
		GROUP BY
			e.code, e.title, e.allowed_for, COALESCE(e.allowed_roles, ARRAY[]::TEXT[]), COALESCE(e.allowed_users, ARRAY[]::TEXT[]), COALESCE(e.allowed_campuses, ARRAY[]::TEXT[]), e.status
`
	queryGetEventAndInstanceByCodes = `
		SELECT
			e.id AS event_id,
			e.code AS event_code,
			e.title AS event_title,
			e.category AS event_category,
			e.topics AS event_topics,
			e.description AS event_description,
			e.terms_and_conditions AS event_terms_and_conditions,
			e.image_links AS event_image_links,
			e.redirect_url AS event_redirect_url,
			e.created_by AS event_created_by,
			e.location_type AS event_location_type,
			e.location_offline_venue AS event_location_offline_venue,
			e.location_online_link AS event_location_online_link,
			e.visibility AS event_visibility,
			e.allowed_community_ids AS event_allowed_community_ids,
			e.allowed_user_types AS event_allowed_user_types,
			e.allowed_roles AS event_allowed_roles,
			e.allowed_campuses AS event_allowed_campuses,
			e.organizer_community_ids AS event_organizer_community_ids, e.is_recurring AS event_is_recurring,
			e.start_at AS event_start_at,
			e.end_at AS event_end_at,
			e.post_details AS event_post_details,
			e.status AS event_status,
			ei.id AS instance_id,
			ei.code AS instance_code,
			ei.title AS instance_title,
			ei.description AS instance_description,
			ei.parent_identifier_fields,
			ei.child_identifier_fields,
			ei.enforce_community_id AS instance_enforce_community_id,
			ei.enforce_uniqueness AS instance_enforce_uniqueness,
			ei.enforce_self_registration AS instance_enforce_self_registration,
			ei.methods AS instance_methods,
			ei.flow AS instance_flow,
			ei.start_at AS instance_start_at,
			ei.end_at AS instance_end_at,
			ei.register_start_at AS instance_register_start_at,
			ei.register_end_at AS instance_register_end_at,
			ei.verify_start_at AS instance_verify_start_at,
			ei.verify_end_at AS instance_verify_end_at,
			ei.timezone AS instance_timezone,
			ei.location_type AS instance_location_type,
			ei.location_offline_venue AS instance_location_offline_venue,
			ei.location_online_link AS instance_location_online_link,
			ei.quota_per_user AS instance_quota_per_user,
			ei.capacity AS instance_capacity,
			ei.post_details AS instance_post_details,
			ei.status AS instance_status
		FROM
			events e
		JOIN
			event_instances ei ON e.code = ei.event_code
		WHERE
			e.code = ? AND ei.code = ?
	`

	queryCheckEventBySlug       = "SELECT EXISTS (SELECT 1 FROM events WHERE slug = ?)"
	queryCheckEventByCodeOrSlug = "SELECT EXISTS (SELECT 1 FROM events WHERE code = ? OR slug = ?)"
)

// buildGetAllEventsQuery constructs a SQL query and its arguments dynamically based on the provided parameters.
// This allows for flexible filtering of events based on user permissions and other criteria.
func buildGetAllEventsQuery(params models.GetAllEventsParams) (string, []interface{}) {
	// 'conditions' will hold the main WHERE clauses, joined by AND.
	var conditions []string
	// 'args' will hold the values for the prepared statement placeholders (?).
	var args []interface{}
	// 'accessConditions' will hold permission-related checks, which will be joined by OR.
	var accessConditions []string

	// Start with the base query to select all events.
	query := queryGetAllEvents

	// Check if the user's roles should be used to filter events.
	if len(params.AllowedRoles) > 0 {
		// The '&&' operator in PostgreSQL checks for array overlap. This finds events
		// where at least one of the event's allowed roles matches the user's roles.
		accessConditions = append(accessConditions, "e.allowed_roles && ?")
		args = append(args, pq.Array(params.AllowedRoles))
	}

	// Check if the user's community IDs should be used to filter events.
	if len(params.AllowedCommunityIDs) > 0 {
		accessConditions = append(accessConditions, "e.allowed_community_ids && ?")
		args = append(args, pq.Array(params.AllowedCommunityIDs))
	}

	// Check if the user's campus affiliations should be used to filter events.
	if len(params.AllowedCampuses) > 0 {
		accessConditions = append(accessConditions, "e.allowed_campuses && ?")
		args = append(args, pq.Array(params.AllowedCampuses))
	}

	// Check if the user's types (e.g., 'volunteer', 'general') should be used to filter events.
	if len(params.AllowedUserTypes) > 0 {
		accessConditions = append(accessConditions, "e.allowed_user_types && ?")
		args = append(args, pq.Array(params.AllowedUserTypes))
	}

	accessConditions = append(accessConditions, "e.visibility = 'public'")

	// If any access control conditions were added, group them together with 'OR'.
	// This means a user can see an event if they match on EITHER role, OR community, OR campus, etc.
	// This entire group is then treated as a single condition in the main 'conditions' slice.
	if len(accessConditions) > 0 {
		conditions = append(conditions, "("+strings.Join(accessConditions, " OR ")+")")
	}

	// Add a filter for event visibility if specified (e.g., 'public', 'private').
	if params.Visibility != "" {
		conditions = append(conditions, "e.visibility = ?")
		args = append(args, params.Visibility)
	}

	// Add a filter for the event status if specified (e.g., 'active', 'draft').
	if params.Status != "" {
		conditions = append(conditions, "e.status = ?")
		args = append(args, params.Status)
	}

	// Add a filter for the event title if specified.
	if params.Title != "" {
		conditions = append(conditions, "e.title ILIKE ?")
		args = append(args, "%"+params.Title+"%")
	}

	// Apply a time-based filter that behaves differently based on the user's type.
	now := common.Now()
	// This checks if the user is NOT an "internal" or "cool" user (i.e., they are a general user).
	if !common.ContainsValueInModel(params.UserTypes, func(userType models.UserType) bool {
		return userType.Category == "internal" || userType.Category == "cool"
	}) {
		// For general users, show events that have started anytime in the past up to one year from now.
		// This allows them to see a history of events but not ones scheduled too far in the future.
		oneYearFromNow := now.AddDate(1, 0, 0)
		conditions = append(conditions, "e.start_at <= ?")
		args = append(args, oneYearFromNow)
	} else {
		// For privileged users ("internal" or "cool"), show only upcoming or currently active events.
		conditions = append(conditions, "e.start_at >= ?")
		args = append(args, now)
	}

	// If any conditions were added, append them to the base query with a WHERE clause, joined by AND.
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += `
		GROUP BY
			e.code, e.title, e.topics, e.location_type, e.visibility, e.allowed_roles, e.allowed_community_ids, e.allowed_campuses, e.is_recurring, e.recurrence, e.start_at, e.end_at, e.status, e.image_links
		ORDER BY
			e.start_at DESC;
	`

	return query, args
}

func BuildQueryGetEventWithInstancesByCodeOrSlug(code string, slug string) (string, string) {
	if code != "" && slug == "" {
		return fmt.Sprintf(queryGetEventWithInstancesByCode, "code"), code
	}
	return fmt.Sprintf(queryGetEventWithInstancesByCode, "slug"), slug
}
