package util

import (
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func ValidateMagicNumber(method, beneficiaryAccount string) (bool, error) {
	switch beneficiaryAccount {
	case "999966660004":
		return true, constant.ErrDuplicateDisbursementReferenceId
	case "999966660007":
		if method == http.MethodGet {
			return true, constant.ErrInsufficientBalance
		}
	case "999966660008": // New magic number for timeout simulation
		time.Sleep(5 * time.Second)
		return true, constant.ErrTimeout

	}
	return false, nil
}
