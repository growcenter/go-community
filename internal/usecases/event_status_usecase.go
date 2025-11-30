package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/models"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/repositories/pgsql"
	"time"
)

type EventStatusUsecase interface {
	DefineAvailabilityStatus(ctx context.Context, event interface{}) (string, error)
}

type eventStatusUsecase struct {
	r pgsql.PostgreRepositories
}

func NewEventStatusUsecase(r pgsql.PostgreRepositories) *eventStatusUsecase {
	return &eventStatusUsecase{r: r}
}

func (esu *eventStatusUsecase) DefineAvailabilityStatus(ctx context.Context, event interface{}) (string, error) {
	// Define a struct to hold the extracted fields
	type eventFields struct {
		totalRemainingSeats  int
		totalSeats           int
		eventRegisterStartAt time.Time
		eventRegisterEndAt   time.Time
		methods              []string
	}

	// Extract fields based on event type
	var fields eventFields

	switch e := event.(type) {
	case models.InstanceDetailDBOutput:
		count, err := esu.r.EventAttendance.CountByInstanceCode(ctx, e.Code)
		if err != nil {
			return "", err
		}

		fields = eventFields{
			totalRemainingSeats:  e.Capacity - int(count),
			totalSeats:           e.Capacity,
			eventRegisterStartAt: e.RegisterStartAt,
			eventRegisterEndAt:   e.RegisterEndAt,
			methods:              e.Methods,
		}

	case models.EventInstance:
		count, err := esu.r.EventAttendance.CountByInstanceCode(ctx, e.Code)
		if err != nil {
			return "", err
		}

		fields = eventFields{
			totalRemainingSeats:  e.Capacity - int(count),
			totalSeats:           e.Capacity,
			eventRegisterStartAt: e.RegisterStartAt,
			eventRegisterEndAt:   e.RegisterEndAt,
			methods:              e.Methods,
		}

	case *models.EventInstance:
		count, err := esu.r.EventAttendance.CountByInstanceCode(ctx, e.Code)
		if err != nil {
			return "", err
		}

		fields = eventFields{
			totalRemainingSeats:  e.Capacity - int(count),
			totalSeats:           e.Capacity,
			eventRegisterStartAt: e.RegisterStartAt,
			eventRegisterEndAt:   e.RegisterEndAt,
			methods:              e.Methods,
		}

	default:
		// Return a default or error if the type is not recognized
		return "", errorgen.Error(errorgen.ErrInvalidInput)
	}

	// Determine availability status based on extracted fields
	// This switch statement evaluates a series of conditions in order to determine the correct availability status for an event instance.
	// The order of these cases is important as it defines the priority of the statuses.
	switch {
	// Case 1: Walk-in Event.
	// If the instance has no capacity (totalSeats is 0) and no registration methods are defined,
	// it's considered a walk-in event where no pre-registration is required.
	case fields.totalSeats == 0 && len(fields.methods) == 0:
		return models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_WALKIN], errorgen.Error(fmt.Errorf("walk-in event"))

	// Case 2: Available (for events without registration).
	// This handles a specific edge case where an event might have a capacity for internal tracking,
	// but no formal registration methods. Even if the tracked attendance meets or exceeds capacity,
	// the event is still shown as 'available' because there's no system to prevent more people from showing up.
	case fields.totalRemainingSeats <= 0 && len(fields.methods) == 0:
		return models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_AVAILABLE], nil

	// Case 3: Full Capacity.
	// If the instance has a defined capacity (totalSeats > 0) and registration methods,
	// and all seats have been taken (totalRemainingSeats <= 0), the event is marked as 'full'.
	case fields.totalRemainingSeats <= 0 && len(fields.methods) > 0 && fields.totalSeats > 0:
		return models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL], errorgen.Error(fmt.Errorf("capacity exceeded"))

	// Case 4: Registration Not Yet Open.
	// If the current time is before the official registration start time, the event is marked as 'soon'.
	case common.Now().Before(fields.eventRegisterStartAt.In(common.GetLocation())):
		return models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON], errorgen.Error(fmt.Errorf("registration not yet open"))

	// Case 5: Registration Closed.
	// If the current time is after the registration period has ended, the event becomes 'unavailable'.
	case common.Now().After(fields.eventRegisterEndAt.In(common.GetLocation())):
		return models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE], errorgen.Error(fmt.Errorf("registration closed"))

	// Default Case: Available.
	// If none of the above conditions are met, the event is considered 'available' for registration.
	default:
		return models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_AVAILABLE], nil
	}
}
