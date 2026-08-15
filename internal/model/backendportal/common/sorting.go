package commonModel

import (
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func ValidateSortOrder(order string) error {
	order = strings.ToUpper(order)
	switch order {
	case constant.SortOrderAsc, constant.SortOrderDesc:
		return nil
	default:
		return constant.ErrInvalidSortOrder
	}
}
