package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

// UpdatePayment updates an existing payment in the system with new details. but its only for VA
//
// This function takes a payment update request and updates the corresponding payment
// in the database. It performs various validations, including:
// - Checking if the payment exists
// - Verifying the payment status is allowed for updates
// - Validating customer information
// - Ensuring merchant ID matches
// - Validating virtual account details based on transaction type
func (s *PaymentService) UpdatePayment(ctx context.Context, req *paymentModel.PaymentUpdateRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/UpdatePayment")
	defer segment.End()

	var (
		paymentResponse        *paymentModel.PaymentResponse
		paymentItemsResponse   []paymentModel.PaymentResponseItem
		paymentVAResponse      *paymentModel.PaymentVirtualAccountResponse
		paymentReferenceId     = ""
		snapCoreResp           snapCoreModel.CreateVirtualAccountResponseData
		paymentRequestCustomer paymentModel.PaymentRequestCustomer
		customerReq            paymentModel.PaymentRequestCustomer
		customer               *customerModel.Customer
		isSnap                 bool
		minAmount              *snapCoreModel.Amount
		maxAmount              *snapCoreModel.Amount
		totalAmountSnapReq     *snapCoreModel.Amount
	)

	// Find payment first
	payment, err := s.paymentRepo.GetPaymentById(ctx, req.PaymentId)
	if err != nil {
		s.logger.Error(ctx, "error when getting payment", logger.Error(err))
		return nil, pkgError.New(httpResponse.HttpErrInternal, err)
	}

	if payment == nil {
		s.logger.Error(ctx, "payment not found", logger.String("id", req.PaymentId))
		return nil, pkgError.New(httpResponse.HttpErrNotFound, fmt.Errorf("payment not found"))
	}

	amount := payment.Amount
	totalAmount := payment.TotalAmount

	if !util.InArray(allowedStatusForPaymentUpdate, payment.Status) {
		s.logger.Error(ctx, "payment already processed", logger.String("id", req.PaymentId))
		return nil, pkgError.New(httpResponse.HttpErrRequest, fmt.Errorf("payment already processed"))
	}

	// get customer detail
	customer, err = s.customerRepo.FindCustomerById(ctx, payment.CustomerID)
	if err != nil {
		s.logger.Error(ctx, "error when get customer data by id", logger.Error(err))
		return nil, pkgError.New(httpResponse.HttpErrInternal, err)
	}

	// get customer info from additional info request
	jsonAdditionalInfo, _ := json.Marshal(req.AdditionalInfo)
	json.Unmarshal(jsonAdditionalInfo, &struct {
		Customer *paymentModel.PaymentRequestCustomer `json:"customer"`
	}{
		Customer: &customerReq,
	})
	// validate customer id
	customerFromReq, _ := s.customerRepo.FindCustomerById(ctx, customerReq.CustomerID)
	if customerFromReq != nil {
		customer = customerFromReq
	}
	firstName, lastName := customerModel.FullNameToFirstNameAndLastName(customerReq.Name)
	if customer != nil {
		if customerReq.Name != "" {
			customer.FirstName = firstName
			customer.LastName = lastName
		}

		if req.AccountName != "" {
			firstName, lastName = customerModel.FullNameToFirstNameAndLastName(req.AccountName)
			customer.FirstName = firstName
			customer.LastName = lastName
		}

		if req.AccountEmail != "" {
			customer.Email = req.AccountEmail
		}

		if req.AccountPhone != "" {
			customer.PhoneNumber = req.AccountPhone
		}
		// convert paymentRequestCustomer
		paymentRequestCustomer = paymentModel.PaymentRequestCustomer{
			CustomerID: customer.UUID,
			Name:       customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName),
			Email:      customer.Email,
			Phone:      customer.PhoneNumber,
			Metadata:   nil,
		}
	} else {
		customer = &customerModel.Customer{}
	}

	// check if customer merchant id is same with request merchant id
	if payment.MerchantID != req.MerchantId {
		s.logger.Error(ctx, "merchant id not match", logger.Error(fmt.Errorf("payment not found, merchant id not match")))
		return nil, pkgError.New(httpResponse.HttpErrRequest, fmt.Errorf("payment not found"))
	}

	// Get payment_method by id
	paymentMethod, err := s.paymentMethodRepo.GetPaymentMethodById(ctx, payment.PaymentMethodID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method data by id", logger.Error(err))
		return nil, pkgError.New(httpResponse.HttpErrInternal, err)
	}

	// Get payment items if bill details is empty
	if len(req.BillDetails) == 0 {
		paymentItems, err := s.paymentRepo.GetPaymentItemsByPaymentId(ctx, payment.UUID)
		if err != nil {
			s.logger.Error(ctx, "error when get payment items data by payment id", logger.Error(err))
			return nil, pkgError.New(httpResponse.HttpErrInternal, err)
		}

		for _, item := range paymentItems {
			paymentItemsResponse = append(paymentItemsResponse, *item.ToPaymentResponseItem())
		}
	} else {
		for _, item := range req.BillDetails {
			value, _ := decimal.NewFromString(item.BillAmount.Value)
			paymentItemsResponse = append(paymentItemsResponse, paymentModel.PaymentResponseItem{
				ItemID: item.BillNo,
				Name:   item.BillName,
				Amount: paymentModel.Amount{
					Value:    value,
					Currency: item.BillAmount.Currency,
				},
				Description: item.BillDescription.Indonesia,
			})
		}
	}

	// marshal metadata to json
	jsonData, errMarshal := json.Marshal(payment.Metadata)
	if errMarshal != nil {
		s.logger.Error(ctx, "error when marshal payment metadata", logger.Error(errMarshal))
		return nil, pkgError.New(httpResponse.HttpErrInternal, errMarshal)
	}

	// unmarshal metadata to snapCoreResp
	json.Unmarshal(jsonData, &struct {
		SnapCore interface{} `json:"snapCore"`
		IsSnap   *bool       `json:"isSnap"`
	}{
		SnapCore: &snapCoreResp,
		IsSnap:   &isSnap,
	})

	// TODO: Calculate fee from config later, then fill it in fee column
	fee := decimal.NewFromInt(0)

	if req.TotalAmount != nil {
		amount = req.TotalAmount.Value
		totalAmount = amount.Sub(fee)
	}

	// Validate amount and expired date
	vaTrxType := snapCoreModel.FindVaTrxTypeByCriteria(snapCoreResp.IsClosedAmount, snapCoreResp.IsSingleUse)
	if vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC && req.ExpiredAt != nil && req.ExpiredAt.Before(time.Now()) {
		return nil, pkgError.New(httpResponse.HttpErrRequest, errors.New("expiredDate is not allowed to be less than current time"))
	}
	if (vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC || vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC) && req.TotalAmount != nil && req.TotalAmount.Value.Cmp(decimal.NewFromInt(constantPayment.VIRTUAL_ACCOUNT_MINIMUM_AMOUNT)) < 0 {
		return nil, pkgError.New(httpResponse.HttpErrRequest, errors.New("totalAmount is not allowed to be less than 10000 for type CLOSED_DYNAMIC payment"))
	}

	if vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC && req.MinAmount != nil && req.MaxAmount != nil && req.MinAmount.Value.Cmp(req.MaxAmount.Value) > 0 {
		return nil, pkgError.New(httpResponse.HttpErrRequest, errors.New("minAmount is not allowed to be greater than maxAmount for type OPEN_STATIC payment"))
	}

	// Only update for open static VA
	if vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC {
		if req.MinAmount != nil {
			minAmount = &snapCoreModel.Amount{
				Value:    req.MinAmount.Value.String(),
				Currency: req.MinAmount.Currency,
			}
		}

		if req.MaxAmount != nil {
			maxAmount = &snapCoreModel.Amount{
				Value:    req.MaxAmount.Value.String(),
				Currency: req.MaxAmount.Currency,
			}
		}
	}

	// Only update for closed dynamic and closed static VA
	if vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC || vaTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC {
		if req.TotalAmount != nil {
			totalAmountSnapReq = &snapCoreModel.Amount{
				Value:    req.TotalAmount.Value.String(),
				Currency: req.TotalAmount.Currency,
			}
		}
	}

	// Call snapCore to update payment
	updateReq := snapCoreModel.UpdateVirtualAccountRequest{
		Number:       snapCoreResp.VirtualAccountNo,
		TotalAmount:  totalAmountSnapReq,
		ExpiredAt:    req.ExpiredAt,
		AccountName:  req.AccountName,
		AccountEmail: req.AccountEmail,
		AccountPhone: req.AccountPhone,
		BillDetails:  req.BillDetails,
		CustomerNo:   customerReq.CustomerID,
		MinAmount:    minAmount,
		MaxAmount:    maxAmount,
	}
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    payment.UUID,
		ReferenceId: payment.MerchantID,
		From:        serviceName,
	})

	resUpdate, errUpdateVA := s.snapCoreRepo.UpdateVirtualAccount(ctx, updateReq)
	if errUpdateVA != nil {
		s.logger.Error(ctx, "error when update virtual account", logger.Error(errUpdateVA))
		return nil, pkgError.New(httpResponse.HttpErrInternal, errUpdateVA)
	}

	if resUpdate.ExpiredAt == nil {
		resUpdate.ExpiredAt = &time.Time{}
	}

	if resUpdate.MinAmount == nil {
		resUpdate.MinAmount = &snapCoreModel.Amount{}
		req.MinAmount = nil
	}

	if resUpdate.MaxAmount == nil {
		resUpdate.MaxAmount = &snapCoreModel.Amount{}
		req.MaxAmount = nil
	}

	snapCoreResp.AccountName = resUpdate.AccountName
	snapCoreResp.VirtualAccountNo = resUpdate.Number
	snapCoreResp.ExpiredAt = *resUpdate.ExpiredAt
	snapCoreResp.TotalAmount = resUpdate.TotalAmount
	snapCoreResp.MinAmount = *resUpdate.MinAmount
	snapCoreResp.MaxAmount = *resUpdate.MaxAmount

	// Begin Tx
	ctx, err = s.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	// its safe to assume that the payment is exist
	(*payment.Metadata)["snapCore"] = snapCoreResp
	(*payment.Metadata)["isSnap"] = isSnap
	payment.TotalAmount = totalAmount
	metadataBytes, _ := json.Marshal(payment.Metadata)

	// Update customer
	if customer.UUID != "" {
		s.customerRepo.Update(ctx, customer)
	}

	// Update payment items
	if len(req.BillDetails) != 0 {
		s.paymentRepo.UpdatePaymentItemsFromPaymentResponseItem(ctx, payment.UUID, paymentItemsResponse)
	}

	// Update payment
	errUpdate := s.paymentRepo.UpdatePayment(ctx, payment.UUID, amount, totalAmount, string(metadataBytes), customer.UUID, *req.ExpiredAt)
	if errUpdate != nil {
		s.logger.Error(ctx, "error when updating payment", logger.Error(errUpdate))
		return nil, pkgError.New(httpResponse.HttpErrInternal, errUpdate)
	}

	// Commit Tx
	if err = s.paymentRepo.CommitTransaction(ctx); err != nil {
		return nil, err
	}

	resAmount := paymentModel.Amount{
		Value:    totalAmount,
		Currency: payment.Currency,
	}

	paymentVAResponse = &paymentModel.PaymentVirtualAccountResponse{
		Issuer:                snapCoreResp.Acquirer,
		VirtualAccountTrxType: snapCoreModel.FindVaTrxTypeByCriteria(snapCoreResp.IsClosedAmount, snapCoreResp.IsSingleUse),
		VirtualAccountNumber:  snapCoreResp.VirtualAccountNo,
		VirtualAccountName:    snapCoreResp.AccountName,
		ExpiredDate:           &snapCoreResp.ExpiredAt,
		MinAmount:             req.MinAmount,
		MaxAmount:             req.MaxAmount,
	}

	if payment.ReferenceID != nil {
		paymentReferenceId = *payment.ReferenceID
	}

	paymentResponse = &paymentModel.PaymentResponse{
		UUID:            payment.UUID,
		MerchantID:      payment.MerchantID,
		ReferenceID:     paymentReferenceId,
		Customer:        &paymentRequestCustomer,
		Status:          payment.Status,
		TotalAmount:     &resAmount,
		PaymentMethod:   paymentMethod.Type,
		VirtualAccount:  paymentVAResponse,
		PaymentItems:    &paymentItemsResponse,
		TransactionDate: &payment.CreatedAt,
		LastUpdateDate:  &payment.UpdatedAt,
	}

	if !isSnap {
		paymentResponse.LastUpdateDate = nil
	}

	return paymentResponse, nil
}

func (ps *PaymentService) UpdatePaymentMetadataById(ctx context.Context, paymentID string, metadata paymentModel.UpdatePaymentMetadataRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/UpdatePaymentMetadataById")
	defer segment.End()

	if paymentID == "" {
		return fmt.Errorf("payment ID is required")
	}

	err := ps.paymentRepo.UpdatePaymentMetadataById(ctx, paymentID, metadata)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error updating payment metadata for payment %s", paymentID), logger.Error(err))
		return err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully updated payment metadata for payment %s", paymentID))
	return nil
}
