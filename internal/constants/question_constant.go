package constants

type QuestionType string

const (
	QuestionTypeShortText  QuestionType = "shortText"
	QuestionTypeLongText   QuestionType = "longText"
	QuestionTypeSingle     QuestionType = "singleChoice"
	QuestionTypeMultiple   QuestionType = "multipleChoice"
	QuestionTypeDate       QuestionType = "date"
	QuestionTypeTime       QuestionType = "time"
	QuestionTypeEmail      QuestionType = "email"
	QuestionTypePhone      QuestionType = "phone"
	QuestionTypeEmailPhone QuestionType = "emailPhone"
	QuestionTypeNumber     QuestionType = "number"
	QuestionTypeCampus     QuestionType = "campus"
	QuestionTypeDepartment QuestionType = "department"
	QuestionTypeCool       QuestionType = "cool"
	QuestionTypeLegalId    QuestionType = "legalId"
	QuestionTypeInstagram  QuestionType = "instagram"
)

var MapQuestionType = map[string]QuestionType{
	"shortText":      QuestionTypeShortText,
	"longText":       QuestionTypeLongText,
	"singleChoice":   QuestionTypeSingle,
	"multipleChoice": QuestionTypeMultiple,
	"date":           QuestionTypeDate,
	"time":           QuestionTypeTime,
	"email":          QuestionTypeEmail,
	"phone":          QuestionTypePhone,
	"emailPhone":     QuestionTypeEmailPhone,
	"number":         QuestionTypeNumber,
	"campus":         QuestionTypeCampus,
	"department":     QuestionTypeDepartment,
	"cool":           QuestionTypeCool,
	"legalId":        QuestionTypeLegalId,
	"instagram":      QuestionTypeInstagram,
}
