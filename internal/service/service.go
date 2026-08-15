package service

import (
	"context"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
)

type IDisbursementService interface {
	WPRelease()

	FindByID(ctx context.Context, id string) (*disbursementModel.DisbursementWithTransaction, error)
}
