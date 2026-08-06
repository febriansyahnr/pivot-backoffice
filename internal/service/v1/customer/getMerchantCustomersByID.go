package customerService

import (
	"context"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
)

func (s *CustomerService) GetMerchantCustomersByID(ctx context.Context, merchantId string, customerIds []string) ([]*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/customer/GetMerchantCustomersByID")
	defer segment.End()

	return s.customerRepo.GetMerchantCustomersByID(ctx, merchantId, customerIds)
}
