package constants

type QuestionType string

const (
	QuestionTypeShortText QuestionType = "short_text"
	QuestionTypeLongText  QuestionType = "long_text"
	QuestionTypeSingle    QuestionType = "single_choice"
	QuestionTypeMultiple  QuestionType = "multiple_choice"
	QuestionTypeDate      QuestionType = "date"
	QuestionTypeEmail     QuestionType = "email"
	QuestionTypePhone     QuestionType = "phone"
)
