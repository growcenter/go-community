package usecases

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"go-community/internal/repositories/pgsql"
	"strconv"
	"strings"
	"time"
)

type EventRegistrationRecordUsecase interface {
	Create(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) (response *models.CreateEventRegistrationRecordResponse, err error)
	GetAll(ctx context.Context) (userTypes []models.UserType, err error)
	UpdateStatus(ctx context.Context, requestParam *models.UpdateRegistrationStatusParameter, requestBody *models.UpdateRegistrationStatusRequest, value *models.TokenValues) (response *models.UpdateRegistrationStatusResponse, err error)
	GetAttendance(ctx context.Context, request models.GetEventAttendanceParameter) (detail *models.GetEventAttendanceDetailResponse, list []models.GetEventAttendanceListResponse, err error)
	GetAllCursor(ctx context.Context, params models.GetAllRegisteredCursorParam) (res []models.GetAllRegisteredCursorResponse, total int, err error)
	Download(ctx context.Context, param models.GetDownloadAllRegisteredParam) (data []byte, contentType string, err error)
}

type eventRegistrationRecordUsecase struct {
	r   pgsql.PostgreRepositories
	cfg config.Configuration
}

func NewEventRegistrationRecordUsecase(r pgsql.PostgreRepositories, cfg config.Configuration) *eventRegistrationRecordUsecase {
	return &eventRegistrationRecordUsecase{
		r:   r,
		cfg: cfg,
	}
}

func (erru *eventRegistrationRecordUsecase) Create(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) (response *models.CreateEventRegistrationRecordResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	if err = erru.validateCreate(ctx, request, value); err != nil {
		return nil, err
	}

	return erru.createAtomic(ctx, request, value)
}

func (erru *eventRegistrationRecordUsecase) createAtomic(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) (response *models.CreateEventRegistrationRecordResponse, err error) {
	res := &models.CreateEventRegistrationRecordResponse{}

	var (
		name              string
		registerStatus    string
		communityIdOrigin string
		verifiedAt        sql.NullTime
		updatedBy         string
		registerAt        time.Time
	)

	registerAt, _ = common.ParseStringToDatetime(time.RFC3339, request.RegisterAt, common.GetLocation())

	if request.IsPersonalQR {
		registerStatus = models.MapRegisterStatus[models.REGISTER_STATUS_SUCCESS]
		communityIdOrigin = request.CommunityId
		verifiedAt = sql.NullTime{
			Time:  registerAt,
			Valid: true,
		}
		updatedBy = "user"

		nameRegister, err := erru.r.User.GetUserNameByCommunityId(ctx, request.CommunityId)
		if err != nil {
			return nil, err
		}
		name = nameRegister.Name
	} else {
		registerStatus = models.MapRegisterStatus[models.REGISTER_STATUS_PENDING]
		communityIdOrigin = value.CommunityId
		verifiedAt = sql.NullTime{
			Valid: false,
		}
		updatedBy = ""
		name = common.StringTrimSpaceAndUpper(request.Name)
	}

	err = erru.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		countTotalRegistrants := 1 + len(request.Registrants)
		var register = make([]models.EventRegistrationRecord, 0, countTotalRegistrants)
		instance, err := r.EventInstance.GetSeatsNamesByCode(ctx, request.InstanceCode)
		if err != nil {
			return err
		}

		if instance == nil {
			return models.ErrorDataNotFound
		}

		main := models.EventRegistrationRecord{
			ID:                uuid.New(),
			Name:              name,
			Identifier:        request.Identifier,
			CommunityId:       request.CommunityId,
			EventCode:         request.EventCode,
			InstanceCode:      request.InstanceCode,
			IdentifierOrigin:  request.Identifier,
			CommunityIdOrigin: communityIdOrigin,
			Status:            registerStatus,
			RegisteredAt:      registerAt,
			VerifiedAt:        verifiedAt,
			Description:       request.Description,
			UpdatedBy:         updatedBy,
		}

		register = append(register, main)

		for _, registrant := range request.Registrants {
			register = append(register, models.EventRegistrationRecord{
				ID:                uuid.New(),
				Name:              common.StringTrimSpaceAndUpper(registrant.Name),
				EventCode:         request.EventCode,
				InstanceCode:      request.InstanceCode,
				IdentifierOrigin:  request.Identifier,
				CommunityIdOrigin: communityIdOrigin,
				Description:       request.Description,
				Status:            registerStatus,
				RegisteredAt:      registerAt,
			})
		}

		if err = r.EventRegistrationRecord.BulkCreate(ctx, &register); err != nil {
			return err
		}

		instance.BookedSeats += countTotalRegistrants

		if instance.TotalSeats != 0 {
			if instance.BookedSeats > instance.TotalSeats {
				return models.ErrorRegisterQuotaNotAvailable
			}

			if (instance.TotalRemainingSeats - instance.BookedSeats) <= 0 {
				return models.ErrorRegisterQuotaNotAvailable
			}
		}

		if request.IsPersonalQR {
			instance.ScannedSeats += countTotalRegistrants
			if err = r.EventInstance.UpdateSeatsByCode(ctx, request.InstanceCode, instance); err != nil {
				return err
			}
		} else {
			if err = r.EventInstance.UpdateBookedSeatsByCode(ctx, request.InstanceCode, instance); err != nil {
				return err
			}
		}

		registrantRes := make([]models.CreateOtherEventRegistrationRecordResponse, len(register))
		for i, p := range register {
			registrantRes[i] = models.CreateOtherEventRegistrationRecordResponse{
				Type:   models.TYPE_EVENT_REGISTRATION_RECORD,
				ID:     p.ID,
				Name:   p.Name,
				Status: p.Status,
			}
		}

		res = &models.CreateEventRegistrationRecordResponse{
			Type:             models.TYPE_EVENT_REGISTRATION_RECORD,
			ID:               main.ID,
			Status:           registerStatus,
			Name:             main.Name,
			Identifier:       main.Identifier,
			CommunityID:      main.CommunityId,
			EventCode:        request.EventCode,
			EventTitle:       instance.EventTitle,
			InstanceCode:     request.InstanceCode,
			InstanceTitle:    instance.EventInstanceTitle,
			TotalRegistrants: countTotalRegistrants,
			Description:      request.Description,
			RegisterAt:       registerAt,
			Registrants:      registrantRes[1:],
		}

		return nil
	})
	return res, err
}

func (erru *eventRegistrationRecordUsecase) validateCreate(ctx context.Context, request *models.CreateEventRegistrationRecordRequest, value *models.TokenValues) error {
	if request.EventCode != request.InstanceCode[:7] {
		return models.ErrorMismatchFields
	}

	if request.Identifier == "" && request.CommunityId == "" {
		return models.ErrorIdentifierCommunityIdEmpty
	}

	if request.IsPersonalQR {
		if request.CommunityId == "" {
			return models.ErrorInvalidInput
		}

		userExist, err := erru.r.User.GetByCommunityId(ctx, request.CommunityId)
		if err != nil {
			return err
		}

		if &userExist == nil {
			return models.ErrorDataNotFound
		}

	}

	countTotalRegistrants := 1 + len(request.Registrants)
	if request.IsPersonalQR && countTotalRegistrants > 1 {
		return models.ErrorQRForMoreThanOneRegister
	}

	event, err := erru.r.Event.GetOneByCode(ctx, request.EventCode)
	if err != nil {
		return err
	}

	eventAvailableStatus, err := models.DefineAvailabilityStatus(event)
	if err != nil {
		return err
	}

	registerAt, _ := common.ParseStringToDatetime(time.RFC3339, request.RegisterAt, common.GetLocation())
	switch {
	case event.EventCode == "" || event.EventStatus != models.MapStatus[models.STATUS_ACTIVE]:
		return models.ErrorDataNotFound
	case request.EventCode != event.EventCode:
		return models.ErrorEventNotValid
	//case common.Now().Before(event.EventRegisterStartAt.In(common.GetLocation())):
	//	return models.ErrorCannotRegisterYet
	//case common.Now().After(event.EventRegisterEndAt.In(common.GetLocation())):
	//	return models.ErrorRegistrationTimeDisabled
	case registerAt.Before(event.EventRegisterStartAt.In(common.GetLocation())):
		return models.ErrorCannotRegisterYet
	case registerAt.After(event.EventRegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	//case request.IsPersonalQR && event.EventAllowedFor != "public":
	//	isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, value.Roles)
	//	isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUsers, value.UserTypes)
	//	fmt.Println(isAllowedUsers, isAllowedRoles)
	//	if !isAllowedRoles && !isAllowedUsers {
	//		return models.ErrorForbiddenRole
	//	}
	case !request.IsPersonalQR && event.EventAllowedFor != "public":
		userExist, err := erru.r.User.GetByCommunityId(ctx, request.CommunityId)
		if err != nil {
			return err
		}

		if &userExist == nil {
			return models.ErrorDataNotFound
		}

		isAllowedRoles := common.CheckOneDataInList(event.EventAllowedRoles, userExist.Roles)
		isAllowedUsers := common.CheckOneDataInList(event.EventAllowedUsers, userExist.UserTypes)
		if !isAllowedRoles && !isAllowedUsers {
			return models.ErrorForbiddenRole
		}
	case eventAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE]:
		return models.ErrorEventNotAvailable
	case eventAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL]:
		return models.ErrorRegisterQuotaNotAvailable
	case eventAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON]:
		return models.ErrorCannotRegisterYet
	}

	instance, err := erru.r.EventInstance.GetOneByCode(ctx, request.InstanceCode, models.MapStatus[models.STATUS_ACTIVE])
	if err != nil {
		return err
	}

	instanceAvailableStatus, err := models.DefineAvailabilityStatus(instance)
	if err != nil {
		return err
	}

	switch {
	case instance.InstanceCode == "" || instance.InstanceStatus != models.MapStatus[models.STATUS_ACTIVE]:
		return models.ErrorDataNotFound
	case instance.InstanceEventCode != request.EventCode || instance.InstanceEventCode != event.EventCode:
		return models.ErrorEventNotValid
	case instance.InstanceRegisterFlow == models.MapRegisterFlow[models.REGISTER_FLOW_NONE]:
		return models.ErrorNoRegistrationNeeded
	case request.IsPersonalQR && instance.InstanceRegisterFlow == models.MapRegisterFlow[models.REGISTER_FLOW_EVENT]:
		return models.ErrorCannotUsePersonalQR
	//case instanceAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_UNAVAILABLE]:
	//	return models.ErrorEventNotAvailable
	case registerAt.After(event.EventRegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	case registerAt.After(instance.InstanceRegisterEndAt.In(common.GetLocation())):
		return models.ErrorRegistrationTimeDisabled
	case ((instance.TotalRemainingSeats - countTotalRegistrants) <= 0) && instance.InstanceRegisterFlow != models.MapRegisterFlow[models.REGISTER_FLOW_NONE] && event.EventIsRecurring == false && instance.InstanceTotalSeats > 0:
		return models.ErrorRegisterQuotaNotAvailable
	case instanceAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_FULL]:
		//if event.EventAllowedFor != "private" {
		//	return models.ErrorRegisterQuotaNotAvailable
		//}

		return models.ErrorRegisterQuotaNotAvailable
	//case instanceAvailableStatus == models.MapAvailabilityStatus[models.AVAILABILITY_STATUS_SOON]:
	//	return models.ErrorCannotRegisterYet
	case registerAt.Before(event.EventRegisterStartAt.In(common.GetLocation())):
		return models.ErrorCannotRegisterYet
	case instance.InstanceMaxPerTransaction > 0 && countTotalRegistrants > instance.InstanceMaxPerTransaction:
		return models.ErrorExceedMaxSeating
	}

	switch {
	case instance.InstanceIsOnePerAccount:
		countRegistered, err := erru.r.EventRegistrationRecord.CountByCommunityIdOriginAndInstanceCode(ctx, common.StringTrimSpaceAndLower(request.CommunityId), common.StringTrimSpaceAndLower(request.InstanceCode))
		if err != nil {
			return err
		}
		if countRegistered > 0 {
			return models.ErrorEventCanOnlyRegisterOnce
		}
	case instance.InstanceIsOnePerTicket:
		if request.Identifier != "" && request.CommunityId == "" {
			identifierExist, err := erru.r.EventRegistrationRecord.CheckByIdentifierAndInstanceCode(ctx, common.StringTrimSpaceAndLower(request.Identifier), common.StringTrimSpaceAndLower(request.InstanceCode))
			if err != nil {
				return err
			}
			if identifierExist {
				return models.ErrorAlreadyRegistered
			}
		} else if request.Identifier == "" && request.CommunityId != "" {
			communityIdExist, err := erru.r.EventRegistrationRecord.CheckByCommunityIdAndInstanceCode(ctx, request.CommunityId, common.StringTrimSpaceAndLower(request.InstanceCode))
			if err != nil {
				return err
			}
			if communityIdExist {
				return models.ErrorAlreadyRegistered
			}
		} else {
			return models.ErrorIdentifierCommunityIdEmpty
		}

		if len(request.Registrants) > 0 {
			for _, registrant := range request.Registrants {
				nameExist, err := erru.r.EventRegistrationRecord.CheckByNameAndInstanceCode(ctx, common.StringTrimSpaceAndUpper(registrant.Name), common.StringTrimSpaceAndLower(request.InstanceCode))
				if err != nil {
					return err
				}
				if nameExist {
					return models.ErrorAlreadyRegistered
				}
			}
		}
	case instance.InstanceIsOnePerTicket && instance.InstanceIsOnePerAccount:
		countRegistered, err := erru.r.EventRegistrationRecord.CountByCommunityIdOriginAndInstanceCode(ctx, common.StringTrimSpaceAndLower(request.CommunityId), common.StringTrimSpaceAndLower(request.InstanceCode))
		if err != nil {
			return err
		}
		if countRegistered > 0 {
			return models.ErrorEventCanOnlyRegisterOnce
		}
	default:
		countRegistered, err := erru.r.EventRegistrationRecord.CountByIdentifierOriginAndStatus(ctx, common.StringTrimSpaceAndLower(request.Identifier), models.MapRegisterStatus[models.REGISTER_STATUS_PENDING])
		if err != nil {
			return err
		}

		if instance.InstanceMaxPerTransaction > 0 && ((int(countRegistered) + countTotalRegistrants) > instance.InstanceMaxPerTransaction) {
			return models.ErrorExceedMaxSeating
		}
	}

	return nil
}

func (erru *eventRegistrationRecordUsecase) UpdateStatus(ctx context.Context, requestParam *models.UpdateRegistrationStatusParameter, requestBody *models.UpdateRegistrationStatusRequest, value *models.TokenValues) (response *models.UpdateRegistrationStatusResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	allowedRoles := []string{"event-verify-record"}
	allowedUsers := []string{"admin", "superadmin", "usher", "volunteer"}
	switch requestBody.Status {
	case models.MapRegisterStatus[models.REGISTER_STATUS_SUCCESS]:
		isAllowedRoles := common.CheckOneDataInList(allowedRoles, value.Roles)
		isAllowedUsers := common.CheckOneDataInList(allowedUsers, value.UserTypes)
		if !isAllowedRoles && !isAllowedUsers {
			return nil, models.ErrorForbiddenRole
		}
	case models.MapRegisterStatus[models.REGISTER_STATUS_CANCELLED]:
	//case models.MapRegisterStatus[models.REGISTER_STATUS_PENDING]:
	//	return nil, models.ErrorInvalidInput
	default:
		return nil, models.ErrorInvalidInput
	}

	record, err := erru.r.EventRegistrationRecord.GetById(ctx, requestParam.ID)
	if err != nil {
		return nil, err
	}

	if record.ID == uuid.Nil {
		return nil, models.ErrorDataNotFound
	}

	switch record.Status {
	case models.MapRegisterStatus[models.REGISTER_STATUS_SUCCESS]:
		return nil, models.ErrorAlreadyVerified
	case models.MapRegisterStatus[models.REGISTER_STATUS_CANCELLED]:
		return nil, models.ErrorAlreadyCancelled
	case models.MapRegisterStatus[models.REGISTER_STATUS_PENDING]:
	case models.MapRegisterStatus[models.REGISTER_STATUS_PERMIT]:
		event, err := erru.r.Event.GetOneByCode(ctx, record.EventCode)
		if err != nil {
			return nil, err
		}

		if event.EventAllowedFor != "private" {
			return nil, models.ErrorForbiddenStatus
		}

		if requestBody.Reason == "" {
			return nil, models.ErrorReasonEmpty
		}
	default:
		return nil, models.ErrorInvalidInput
	}

	res := models.UpdateRegistrationStatusResponse{}
	err = erru.r.Transaction.Atomic(ctx, func(ctx context.Context, r *pgsql.PostgreRepositories) error {
		record.Status = requestBody.Status
		record.Reason = requestBody.Reason
		verifiedAt, _ := common.ParseStringToDatetime(time.RFC3339, requestBody.UpdatedAt, common.GetLocation())
		record.VerifiedAt = sql.NullTime{Valid: true, Time: verifiedAt}
		record.UpdatedBy = value.CommunityId

		if err := r.EventRegistrationRecord.Update(ctx, record); err != nil {
			return err
		}

		instance, err := r.EventInstance.GetSeatsNamesByCode(ctx, record.InstanceCode)
		if err != nil {
			return err
		}

		if instance == nil {
			return models.ErrorDataNotFound
		}

		switch requestBody.Status {
		case models.MapRegisterStatus[models.REGISTER_STATUS_SUCCESS]:
			instance.ScannedSeats += 1
			if err = r.EventInstance.UpdateScannedSeatsByCode(ctx, record.InstanceCode, instance); err != nil {
				return err
			}
		case models.MapRegisterStatus[models.REGISTER_STATUS_CANCELLED]:
			instance.BookedSeats -= 1
			if err = r.EventInstance.UpdateBookedSeatsByCode(ctx, record.InstanceCode, instance); err != nil {
				return err
			}
		default:
			return models.ErrorInvalidInput
		}

		res = models.UpdateRegistrationStatusResponse{
			Type:          models.TYPE_EVENT_REGISTRATION_RECORD,
			ID:            record.ID,
			Status:        requestBody.Status,
			Name:          record.Name,
			Identifier:    record.Identifier,
			CommunityID:   record.CommunityId,
			EventCode:     record.EventCode,
			EventTitle:    instance.EventTitle,
			InstanceCode:  record.InstanceCode,
			InstanceTitle: instance.EventInstanceTitle,
			UpdatedBy:     value.CommunityId,
			VerifiedAt:    verifiedAt,
		}

		return nil
	})
	return &res, err
}

func (erru *eventRegistrationRecordUsecase) GetAttendance(ctx context.Context, request models.GetEventAttendanceParameter) (detail *models.GetEventAttendanceDetailResponse, list []models.GetEventAttendanceListResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	userExist, err := erru.r.User.CheckByCommunityId(ctx, request.CommunityId)
	if err != nil {
		return nil, nil, err
	}

	if !userExist {
		return nil, nil, models.ErrorUserNotFound
	}

	var year int
	if request.Year == "" {
		year = common.Now().Year()
	}

	startDate := fmt.Sprintf("%d-01-01 00:00:00", year)
	endDate := fmt.Sprintf("%d-12-31 23:59:59", year)

	attendance, err := erru.r.EventRegistrationRecord.GetEventAttendance(ctx, request.CommunityId, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}

	detailRes := models.GetEventAttendanceDetailResponse{
		Type:           models.TYPE_USER,
		CommunityId:    request.CommunityId,
		AttendanceYear: year,
	}

	var listRes []models.GetEventAttendanceListResponse
	for _, i := range attendance {
		var attendancePercentage float64
		if i.TotalInstances > 0 {
			attendancePercentage = float64(i.SuccessCount) / float64(i.TotalInstances) * 100
		} else {
			attendancePercentage = 0.00
		}

		listRes = append(listRes, models.GetEventAttendanceListResponse{
			Type:                 models.TYPE_EVENT_REGISTRATION_RECORD,
			EventCode:            i.EventCode,
			EventTitle:           i.Title,
			AttendanceCount:      i.SuccessCount,
			PermitCount:          i.PermitWithReasonCount,
			AbsenceCount:         i.OtherStatusCount + i.PermitWithoutReasonCount,
			TotalInstances:       i.TotalInstances,
			AttendancePercentage: attendancePercentage,
		})
	}

	return &detailRes, listRes, nil
}

func (erru *eventRegistrationRecordUsecase) GetAllCursor(ctx context.Context, params models.GetAllRegisteredCursorParam) (res []models.GetAllRegisteredCursorResponse, info *models.CursorInfo, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	output, prev, next, total, err := erru.r.EventRegistrationRecord.GetAllWithCursor(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	//output, err := erru.r.EventRegistrationRecord.GetAllWithCursor(ctx, params)
	//if err != nil {
	//	return nil, err
	//}

	var response []models.GetAllRegisteredCursorResponse
	for _, v := range output {
		var isPersonalQr bool
		if v.UpdatedBy == "user" {
			isPersonalQr = true
		}

		var verifiedAt string
		if !v.VerifiedAt.Time.IsZero() {
			verifiedAt = common.FormatDatetimeToString(v.VerifiedAt.Time, time.RFC3339)
		}

		var deletedAt string
		if !v.DeletedAt.Time.IsZero() {
			deletedAt = common.FormatDatetimeToString(v.DeletedAt.Time, time.RFC3339)
		}

		var departmentName string
		if v.Department != "" {
			value, department := erru.cfg.Department[strings.ToLower(v.Department)]
			if !department {
				return nil, nil, models.ErrorDataNotFound
			}
			departmentName = value
		}

		var campusName string
		if v.CampusCode != "" {
			value, campus := erru.cfg.Campus[strings.ToLower(v.CampusCode)]
			if !campus {
				return nil, nil, models.ErrorDataNotFound
			}
			campusName = value
		}

		response = append(response, models.GetAllRegisteredCursorResponse{
			Type:              models.TYPE_EVENT_REGISTRATION_RECORD,
			ID:                v.ID,
			Name:              v.Name,
			Identifier:        v.Identifier,
			CommunityId:       v.CommunityId,
			Email:             v.Email,
			PhoneNumber:       v.PhoneNumber,
			CampusCode:        v.CampusCode,
			CampusName:        campusName,
			CoolId:            v.CoolId,
			CoolName:          v.CoolName,
			DepartmentCode:    v.Department,
			DepartmentName:    departmentName,
			EventCode:         v.EventCode,
			EventName:         v.EventName,
			InstanceCode:      v.InstanceCode,
			InstanceName:      v.InstanceName,
			IdentifierOrigin:  v.IdentifierOrigin,
			CommunityIdOrigin: v.CommunityIdOrigin,
			RegisteredBy:      v.RegisteredBy,
			UpdatedBy:         v.UpdatedBy,
			IsPersonalQr:      isPersonalQr,
			Description:       v.Description,
			Status:            v.Status,
			RegisteredAt:      v.RegisteredAt,
			VerifiedAt:        verifiedAt,
			CreatedAt:         v.CreatedAt,
			UpdatedAt:         v.UpdatedAt,
			DeletedAt:         deletedAt,
		})
	}
	info = &models.CursorInfo{
		PreviousCursor: prev,
		NextCursor:     next,
		TotalData:      total,
	}

	return response, info, nil
}

func (erru *eventRegistrationRecordUsecase) Download(ctx context.Context, param models.GetDownloadAllRegisteredParam) (data []byte, contentType string, fileName string, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	record, err := erru.r.EventRegistrationRecord.Download(ctx, param)
	if err != nil {
		return nil, "", "", err
	}

	fileName = fmt.Sprintf("%s-%s", record[0].EventName, record[0].InstanceName)

	switch param.Format {
	case "csv":
		output, contentType, err := erru.downloadCSV(record)
		if err != nil {
			return nil, "", "", err
		}

		return output, contentType, fmt.Sprintf("%s.csv", fileName), nil
	case "xlsx":
		output, contentType, err := erru.downloadXLSX(record)
		if err != nil {
			return nil, "", "", err
		}

		return output, contentType, fmt.Sprintf("%s.xlsx", fileName), nil
	default:
		return nil, "", "", nil
	}
}

func (erru *eventRegistrationRecordUsecase) downloadXLSX(data []models.GetDownloadAllRegisteredDBOutput) (file []byte, contentType string, err error) {
	f := excelize.NewFile()
	// Create a new Excel file
	sheetName := "Registration-Record"
	err = f.SetSheetName("Sheet1", sheetName)
	if err != nil {
		return nil, "", err
	}

	// Define headers
	headers := []string{"ID", "Name", "Email/Phone Number", "Community ID", "Email", "Phone Number", "Campus", "COOL", "Department", "Event", "Event Instance", "Description", "Is Using QR", "Register At", "Verified At", "Status"}

	// Write headers to the first row
	for i, header := range headers {
		col := fmt.Sprintf("%c1", 'A'+i) // Convert index to column letter (A, B, C, etc.)
		err = f.SetCellValue(sheetName, col, header)
		if err != nil {
			return nil, "", err
		}
	}

	// Write user data
	for i, user := range data {
		var isPersonalQr bool
		if user.UpdatedBy == "user" {
			isPersonalQr = true
		}

		var verifiedAt string
		if !user.VerifiedAt.Time.IsZero() {
			verifiedAt = common.FormatDatetimeToString(user.VerifiedAt.Time, time.RFC3339)
		}

		var departmentName string
		if user.Department != "" {
			value, department := erru.cfg.Department[strings.ToLower(user.Department)]
			if !department {
				return nil, "", models.ErrorDataNotFound
			}
			departmentName = value
		}

		var campusName string
		if user.CampusCode != "" {
			value, campus := erru.cfg.Campus[strings.ToLower(user.CampusCode)]
			if !campus {
				return nil, "", models.ErrorDataNotFound
			}
			campusName = value
		}

		row := i + 2 // Start from row 2 (after headers)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), user.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), user.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user.Identifier)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), user.CommunityId)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), user.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), user.PhoneNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), campusName)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), user.CoolName)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), departmentName)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), user.EventName)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), user.InstanceName)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), user.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), isPersonalQr)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), user.RegisteredAt)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), verifiedAt)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), user.Status)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("error writing to buffer: %w", err)
	}

	return buffer.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

func (erru *eventRegistrationRecordUsecase) downloadCSV(data []models.GetDownloadAllRegisteredDBOutput) (file []byte, contentType string, err error) {
	buffer := bytes.NewBuffer(nil)
	writer := csv.NewWriter(buffer)

	// Define headers
	headers := []string{"ID", "Name", "Email/Phone Number", "Community ID", "Email", "Phone Number", "Campus", "COOL", "Department", "Event", "Event Instance", "Description", "Is Using QR", "Register At", "Verified At", "Status"}
	if err := writer.Write(headers); err != nil {
		return nil, "", err
	}

	// Write user data
	for _, user := range data {
		var isPersonalQr bool
		if user.UpdatedBy == "user" {
			isPersonalQr = true
		}

		var verifiedAt string
		if !user.VerifiedAt.Time.IsZero() {
			verifiedAt = common.FormatDatetimeToString(user.VerifiedAt.Time, time.RFC3339)
		}

		var departmentName string
		if user.Department != "" {
			value, department := erru.cfg.Department[strings.ToLower(user.Department)]
			if !department {
				return nil, "", models.ErrorDataNotFound
			}
			departmentName = value
		}

		var campusName string
		if user.CampusCode != "" {
			value, campus := erru.cfg.Campus[strings.ToLower(user.CampusCode)]
			if !campus {
				return nil, "", models.ErrorDataNotFound
			}
			campusName = value
		}

		row := []string{
			user.ID.String(),
			user.Name,
			user.Identifier,
			user.CommunityId,
			user.Email,
			user.PhoneNumber,
			campusName,
			user.CoolName,
			departmentName,
			user.EventName,
			user.InstanceName,
			user.Description,
			strconv.FormatBool(isPersonalQr),
			user.RegisteredAt.String(),
			verifiedAt,
			user.Status,
		}

		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	return buffer.Bytes(), "text/csv", nil
}
