package constants

// FormType categorises a form by its intended use.
type FormType string

const (
	FormTypeRegistration FormType = "registration"
	FormTypeSurvey       FormType = "survey"
	FormTypeQuiz         FormType = "quiz"
)
