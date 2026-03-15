package constants

type QuestionType string

const (
	QuestionTypeShortText QuestionType = "short_text"
	QuestionTypeLongText  QuestionType = "long_text"
	QuestionTypeSingle    QuestionType = "single_choice"
	QuestionTypeMultiple  QuestionType = "multiple_choice"
	QuestionTypeDate      QuestionType = "date"
	QuestionTypeTime      QuestionType = "time"
	QuestionTypeEmail     QuestionType = "email"
	QuestionTypePhone     QuestionType = "phone"
	QuestionTypeNumber    QuestionType = "number"
)

var MapQuestionType = map[string]QuestionType{
	"short_text":      QuestionTypeShortText,
	"long_text":       QuestionTypeLongText,
	"single_choice":   QuestionTypeSingle,
	"multiple_choice": QuestionTypeMultiple,
	"date":            QuestionTypeDate,
	"time":            QuestionTypeTime,
	"email":           QuestionTypeEmail,
	"phone":           QuestionTypePhone,
	"number":          QuestionTypeNumber,
}
