package vault

import "fmt"

type ErrMissingExpectedElement string

func (e ErrMissingExpectedElement) Error() string {
	return fmt.Sprintf("missing expected '%s' element", string(e))
}

type ErrInvalidAttribute string

func (e ErrInvalidAttribute) Error() string {
	return string(e)
}

type ErrInvalidDataType string

func (e ErrInvalidDataType) Error() string {
	return fmt.Sprintf("'%s' return data type is invalid", string(e))
}
