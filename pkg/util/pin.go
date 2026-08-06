package util

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

// PINs with sequential numbers (e.g., 123456, 654321) or identical numbers (e.g., 111111) are not allowed.
func IsValidPin(pin string) error {
	if len(pin) != 6 {
		return constant.ErrInvalidPINLength
	}

	// identical digits
	isIdentical := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1] {
			isIdentical = false
			break
		}
	}
	if isIdentical {
		return constant.ErrIdenticalPIN
	}

	// sequence digits (increasing or decreasing)
	isIncreasingSequential := true
	isDecreasingSequential := true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			isIncreasingSequential = false
		}
		if pin[i] != pin[i-1]-1 {
			isDecreasingSequential = false
		}
	}
	if isIncreasingSequential || isDecreasingSequential {
		return constant.ErrSequentialPIN
	}

	return nil
}
