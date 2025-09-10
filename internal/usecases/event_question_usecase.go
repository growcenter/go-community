package usecases

import (
	"context"
	"fmt"
	"go-community/internal/common"
	"go-community/internal/constants"
	"go-community/internal/models"
	"go-community/internal/pkg/errorgen"
	"go-community/internal/repositories/pgsql"
)

type EventQuestionUsecase interface {
	Create(ctx context.Context, request models.CreateQuestionRequest) (response []models.CreateQuestionResponse, err error)
}

type eventQuestionUsecase struct {
	r pgsql.PostgreRepositories
}

func NewEventQuestionUsecase(r pgsql.PostgreRepositories) *eventQuestionUsecase {
	return &eventQuestionUsecase{
		r: r,
	}
}

func (equ *eventQuestionUsecase) GetAllByInstanceCode(ctx context.Context, request models.GetAllByInstanceCodeRequest) (response *models.GetAllQuestionByInstanceCodeResponse, err error) {
	defer func() {
		LogService(ctx, err)
	}()

	var parentQuestions []models.InstanceQuestionForm
	var childQuestions []models.InstanceQuestionForm

	instance, err := equ.r.EventInstance.GetByCode(ctx, request.InstanceCode)
	if err != nil {
		return nil, err
	}

	if instance.ID == 0 {
		return nil, errorgen.Error(errorgen.DataNotFound, "instance not found")
	}

	eventAssociation, err := equ.r.FormAssociation.GetByEntityCode(ctx, models.TYPE_EVENT, instance.EventCode)
	if err != nil {
		return nil, err
	}

	instanceAssociation, err := equ.r.FormAssociation.GetByEntityCode(ctx, models.TYPE_EVENT_INSTANCE, instance.Code)
	if err != nil {
		return nil, err
	}

	if (len(eventAssociation) == 0) && (len(instanceAssociation) == 0) {
		return nil, errorgen.Error(errorgen.DataNotFound, "form association not found")
	}

	eventFormCodes, err := common.GetUniqueFieldValuesFromModelUUID(eventAssociation, "FormCode")
	if err != nil {
		return nil, err
	}

	instanceFormCodes, err := common.GetUniqueFieldValuesFromModelUUID(instanceAssociation, "FormCode")
	if err != nil {
		return nil, err
	}

	nameQuestion := models.InstanceQuestionForm{
		Type:         models.TYPE_EVENT_QUESTION,
		Code:         "",
		QuestionType: string(constants.QuestionTypeShortText),
		QuestionText: "Name",
		Rules:        nil,
		Options:      nil,
		IsMandatory:  true,
		Instruction:  []string{"Please enter your name"},
		DisplayOrder: 1,
	}

	parentQuestions = append(parentQuestions, nameQuestion)
	childQuestions = append(childQuestions, nameQuestion)
	parentQuestions = append(parentQuestions, equ.getIdentifierQuestions(instance.ValidateParentIdentifier, instance.ParentIdentifierInput))
	childQuestions = append(childQuestions, equ.getIdentifierQuestions(instance.ValidateChildIdentifier, instance.ChildIdentifierInput))

	formQuestions, err := equ.r.FormQuestion.GetByFormCodes(ctx, common.UUIDsToStrings(common.CombineMapUUID(eventFormCodes, instanceFormCodes)))
	if err != nil {
		return nil, err
	}

	parentDisplayCount := len(parentQuestions)
	childDisplayCount := len(childQuestions)
	for _, formQuestion := range formQuestions {
		question := models.InstanceQuestionForm{
			Type:         models.TYPE_EVENT_QUESTION,
			Code:         formQuestion.FormCode,
			QuestionType: string(formQuestion.Type),
			QuestionText: formQuestion.Text,
			Rules:        formQuestion.Rules,
			Options:      formQuestion.Options,
			IsMandatory:  false,
			DisplayOrder: 1,
		}

		applyFor := common.CheckPresenceOfValue(formQuestion.ApplyFor, "parent", "child")
		mandatoryFor := common.CheckPresenceOfValue(formQuestion.MandatoryFor, "parent", "child")
		question.Instruction = equ.buildQuestionInstruction(constants.QuestionType(formQuestion.Type), formQuestion.Rules)
		switch {
		case applyFor["parent"] && !applyFor["child"]:
			parentDisplayCount++
			question.DisplayOrder = parentDisplayCount
			if mandatoryFor["parent"] {
				question.IsMandatory = true
			}
			parentQuestions = append(parentQuestions, question)
		case applyFor["child"] && !applyFor["parent"]:
			childDisplayCount++
			question.DisplayOrder = childDisplayCount
			if mandatoryFor["child"] {
				question.IsMandatory = true
			}
			childQuestions = append(childQuestions, question)
		case applyFor["parent"] && applyFor["child"]:
			parentDisplayCount++
			question.DisplayOrder = parentDisplayCount
			if mandatoryFor["parent"] {
				question.IsMandatory = true
			}
			parentQuestions = append(parentQuestions, question)
			childDisplayCount++
			question.DisplayOrder = childDisplayCount
			if mandatoryFor["child"] {
				question.IsMandatory = true
			}
			childQuestions = append(childQuestions, question)
		}
	}

	response = &models.GetAllQuestionByInstanceCodeResponse{
		Type:                models.TYPE_EVENT_QUESTION,
		EventCode:           instance.EventCode,
		InstanceCode:        instance.Code,
		InstanceTitle:       instance.Title,
		InstanceDescription: instance.Description,
		ParentQuestion:      parentQuestions,
		ChildQuestion:       childQuestions,
	}

	return response, nil
}

func (equ *eventQuestionUsecase) getIdentifierQuestions(validateIdentifier bool, identifierInput []string) models.InstanceQuestionForm {
	question := models.InstanceQuestionForm{
		Type:         models.TYPE_EVENT_QUESTION,
		Code:         "",
		QuestionType: string(constants.QuestionTypeEmailPhone),
		QuestionText: "Identifier (Email or Phone)",
		Rules:        nil,
		Options:      nil,
		IsMandatory:  false,
		Instruction:  []string{"[OPTIONAL] Please enter your email or phone number. Example: example@mail.com or 081234567890"},
		DisplayOrder: 2,
	}

	if validateIdentifier {
		inputCheck := common.CheckPresenceOfValue(identifierInput, "email", "phone")
		switch {
		case inputCheck["email"] && !inputCheck["phone"]:
			question.QuestionType = string(constants.QuestionTypeEmail)
			question.QuestionText = "Email"
			question.IsMandatory = true
			question.Instruction = []string{"Please enter your email address. Example: example@mail.com"}
			question.DisplayOrder = 2
			return question
		case !inputCheck["email"] && inputCheck["phone"]:
			question.QuestionType = string(constants.QuestionTypePhone)
			question.QuestionText = "Phone"
			question.IsMandatory = true
			question.Instruction = []string{"Please enter your phone number. Example: 081234567890"}
			question.DisplayOrder = 2
			return question
		case inputCheck["email"] && inputCheck["phone"]:
		default:
			question.QuestionType = string(constants.QuestionTypeEmailPhone)
			question.QuestionText = "Identifier (Email or Phone)"
			question.IsMandatory = true
			question.Instruction = []string{"Please enter your email or phone number. Example: example@mail.com or 081234567890"}
			question.DisplayOrder = 2
			return question
		}
	} else {
		return models.InstanceQuestionForm{
			Type:         models.TYPE_EVENT_QUESTION,
			Code:         "",
			QuestionType: string(constants.QuestionTypeEmail),
			QuestionText: "Email",
			Rules:        nil,
			Options:      nil,
			IsMandatory:  true,
			Instruction:  []string{"Please enter your email or phone number. Example: example@mail.com or 081234567890"},
			DisplayOrder: 2,
		}
	}

	return question
}

func (equ *eventQuestionUsecase) buildQuestionInstruction(questionType constants.QuestionType, rules *models.QuestionValidationRules) []string {
	var instructions []string

	if rules == nil {
		return nil
	}

	switch questionType {
	case constants.QuestionTypeShortText, constants.QuestionTypeLongText:
		if rules.MinLength != nil && *rules.MinLength > 0 {
			instructions = append(instructions, fmt.Sprintf("Minimum characters: %d.", *rules.MinLength))
		}
		if rules.MaxLength != nil && *rules.MaxLength > 0 {
			instructions = append(instructions, fmt.Sprintf("Maximum characters: %d.", *rules.MaxLength))
		}
	case constants.QuestionTypeNumber:
		if rules.MinValue != nil && *rules.MinValue > 0 {
			instructions = append(instructions, fmt.Sprintf("Minimum value: %d.", *rules.MinValue))
		}
		if rules.MaxValue != nil && *rules.MaxValue > 0 {
			instructions = append(instructions, fmt.Sprintf("Maximum value: %d.", *rules.MaxValue))
		}
	case constants.QuestionTypeMultiple:
		if rules.MinSelection != nil && *rules.MinSelection > 0 {
			instructions = append(instructions, fmt.Sprintf("Select at least %d option(s).", *rules.MinSelection))
		}
		if rules.MaxSelection != nil && *rules.MaxSelection > 0 {
			instructions = append(instructions, fmt.Sprintf("Select at most %d option(s).", *rules.MaxSelection))
		}
	case constants.QuestionTypeSingle:
		instructions = append(instructions, "Select one option.")
	case constants.QuestionTypeDate:
		instructions = append(instructions, "Select a date.")
	case constants.QuestionTypeTime:
		instructions = append(instructions, "Select a time.")
	case constants.QuestionTypeEmail:
		instructions = append(instructions, "Enter a valid email address. Example: example@mail.com")
	case constants.QuestionTypePhone:
		instructions = append(instructions, "Enter a valid phone number. Example: +6281234567890")
	case constants.QuestionTypeEmailPhone:
		instructions = append(instructions, "Enter a valid email or phone number. Example: example@mail.com or +6281234567890")
	case constants.QuestionTypeCampus:
		instructions = append(instructions, "Select a campus.")
	case constants.QuestionTypeDepartment:
		instructions = append(instructions, "Select a department.")
	case constants.QuestionTypeCool:
		instructions = append(instructions, "Select your COOL.")
	}

	if len(instructions) > 0 {
		return instructions
	}

	return nil
}
