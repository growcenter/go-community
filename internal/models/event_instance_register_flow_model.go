package models

import (
	"regexp"
	"strings"
)

type RegisterFlow int32

const (
	REGISTER_FLOW_PERSONAL RegisterFlow = iota
	REGISTER_FLOW_EVENT
	REGISTER_FLOW_BOTH
	REGISTER_FLOW_NONE
)

const (
	RegisterFlowPersonal = "personal-qr"
	RegisterFlowEvent    = "event-qr"
	RegisterFlowRegister = "register-qr"
	RegisterFlowBoth     = "both-qr"
	RegisterFlowNone     = "none"
)

var (
	MapRegisterFlow = map[RegisterFlow]string{
		REGISTER_FLOW_PERSONAL: RegisterFlowPersonal,
		REGISTER_FLOW_EVENT:    RegisterFlowEvent,
		REGISTER_FLOW_BOTH:     RegisterFlowBoth,
		REGISTER_FLOW_NONE:     RegisterFlowNone,
	}
)

func GetRegisterFlowsFromStringArray(input []string) []string {
	//var registerFlows []string
	//re := regexp.MustCompile(`(?i)(event-qr|personal-qr|both-qr|none)$`)
	//
	//// Loop through each string in the input slice
	//for _, element := range input {
	//	// Use regex to extract the last meaningful part (e.g., event-qr, personal-qr, etc.)
	//	matches := re.FindStringSubmatch(element)
	//
	//	// If there is a match, append it to the result slice
	//	if len(matches) > 0 {
	//		registerFlows = append(registerFlows, strings.ToLower(matches[0])) // Normalize to lowercase
	//	}
	//}

	var result []string
	for _, item := range input {
		trimmed := strings.Trim(item, "()")  // remove parentheses
		parts := strings.Split(trimmed, ",") // split by comma
		if len(parts) == 3 {
			result = append(result, parts[2]) // get the 3rd part (both-qr)
		}
	}

	return result
}

func CountTotalRegisterFlows(input []string) int {
	// Initialize a total count variable
	totalCount := 0

	// Define a regex to extract the QR type
	re := regexp.MustCompile(`(?i)(event-qr|personal-qr|both-qr)$`)

	// Loop through each input string
	for _, element := range input {
		// Match the QR type
		matches := re.FindStringSubmatch(element)

		// Increment the total count if there's a match
		if len(matches) > 0 {
			totalCount++
		}
	}

	return totalCount
}

func RegisterFlowToCount(flow string) int {
	switch flow {
	case MapRegisterFlow[REGISTER_FLOW_PERSONAL]:
		return 1
	case MapRegisterFlow[REGISTER_FLOW_EVENT]:
		return 1
	case MapRegisterFlow[REGISTER_FLOW_BOTH]:
		return 1
	case MapRegisterFlow[REGISTER_FLOW_NONE]:
		return 0
	default:
		return 0
	}
}
