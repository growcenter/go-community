package herrx

import "github.com/jeremygprawira/herr"

var (
	ErrDeprecatedEndpoint = DefStatus("COMMON_DEPRECATED_ENDPOINT", herr.KindNotFound, 410, "This endpoint has been retired. Please update the app.")
	ErrInvalidInput       = Def("COMMON_INVALID_INPUT", herr.KindInvalid, "Something in your submission does not look right. Please review and try again.")
	ErrNotFound           = Def("COMMON_NOT_FOUND", herr.KindNotFound, "We could not find what you are looking for.")
	ErrForbidden          = Def("COMMON_FORBIDDEN", herr.KindForbidden, "You do not have access to do this.")
	ErrUnauthorized       = Def("COMMON_UNAUTHORIZED", herr.KindUnauthorized, "Please sign in to continue.")
	ErrInternal           = Def("COMMON_INTERNAL", herr.KindInternal, "Something went wrong on our side. Please try again.")
)
