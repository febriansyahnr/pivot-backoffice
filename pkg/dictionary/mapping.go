package dictionary

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func GetAPITranslationCodeByError(err string) string {
	switch err {
	case constant.ErrInvalidPassword.Error():
		return TranslationAPIErrInvalidPassword
	case constant.ErrUserNotFound.Error():
		return TranslationAPIErrUserNotFound
	case constant.ErrInvalidValidation.Error():
		return TranslationAPIErrInvalidValidation

	default:
		return ""
	}
}

func GetInternalTranslationCodeByError(err string) string {
	switch err {
	case constant.ErrInvalidPassword.Error():
		return TranslationInternalErrInvalidCredentials
	default:
		return TranslationErrInternal
	}
}
