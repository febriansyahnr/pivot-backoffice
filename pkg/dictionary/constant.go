package dictionary

const (
	EN = "en"
	IN = "in"
)

const (
	// General Translation Code
	TranslationErrInternal                  = "general.error.message.internal"
	TranslationErrValidationRequired        = "general.error.message.validation.required"
	TranslationErrValidationEmail           = "general.error.message.validation.email"
	TranslationErrValidationMin             = "general.error.message.validation.min.slice"
	TranslationErrValidationOneOf           = "general.error.message.validation.oneof"
	TranslationErrValidationNumeric         = "general.error.message.validation.numeric"
	TranslationErrValidationLength          = "general.error.message.validation.length"
	TranslationErrValidationIso8601Datetime = "general.error.message.validation.iso_8601_datetime"
	TranslationErrValidationLuhn            = "general.error.message.validation.luhn"

	// Api Translation Code for Merchant Portal
	TranslationAPIErrInvalidPassword        = "api.error.message.invalid-password"
	TranslationAPIErrUserNotFound           = "api.error.message.user-not-found"
	TranslationAPIErrInvalidValidation      = "api.error.message.invalid-validation"
	TranslationAPIErrDisbursementMinAmount  = "api.error.message.disbursement.min.amount"
	TranslationAPIErrEmailAlreadyRegistered = "api.error.message.email-already-registered"

	// Internal Translation Code for Open API
	TranslationInternalErrInvalidCredentials = "internal.error.message.invalid.password"
)
