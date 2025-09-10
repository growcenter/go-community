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
