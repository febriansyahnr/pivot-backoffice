package paymentService

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func (s *PaymentService) CreatePayment(ctx context.Context, merchantID string, paymentRequest paymentModel.PaymentRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/CreatePayment")
	defer segment.End()

	if paymentRequest.InitiateStatus == "" {
		paymentRequest.InitiateStatus = constant.PAYMENT_STATUS_PENDING
	}

	// create customer, when we don't have customer data
	customer, err := s.findOrCreateCustomer(ctx, paymentRequest.Customer, merchantID)
	if err != nil {
		return nil, err
	}

	if customer != nil {
		paymentRequest.Customer.CustomerID = customer.UUID
	}

	switch paymentRequest.PaymentMethod {
	case constant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		return s.createPaymentUsingVirtualAccount(ctx, merchantID, paymentRequest)

	case constant.PAYMENT_METHOD_QRIS:
		return s.createPaymentUsingQrMpm(ctx, merchantID, paymentRequest)

	default:
		return nil, errors.New("method is not allowed")
	}
}

func (s *PaymentService) createPaymentUsingVirtualAccount(ctx context.Context, merchantID string, paymentRequest paymentModel.PaymentRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/createPaymentUsingVirtualAccount")
	defer segment.End()

	var (
		pendingLedgerID      = uuid.New()
		paymentResponse      paymentModel.PaymentResponse
		paymentItemsResponse []paymentModel.PaymentResponseItem
		minAmount            snapCoreModel.Amount
		maxAmount            snapCoreModel.Amount
		parentMerchant       *merchantModel.Merchant

		// Initiate VA additional info
		additionalInfo = &map[string]interface{}{
			c.ProcessorExternalIDKey: pendingLedgerID.String(),
		}
	)

	derivedMerchantId := merchantID
	virtualAccountReq := paymentRequest.VirtualAccount

	// FeatureFlag check for exist maximum amount and validate for request payment
	defaultMaxAmount := s.getFlagMaximumAmount()
	if !defaultMaxAmount.Equal(decimal.Zero) && paymentRequest.TotalAmount.Value.Cmp(defaultMaxAmount) > 0 {
		return nil, pkgErrors.New(response.HttpErrRequest, c.ErrInvalidAmount)
	}

	// Get Merchant By merchantID
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, c.ErrFindMerchant)
	} else if merchant == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, c.ErrMerchantNotFound)
	}

	if merchant.ParentID.String != "" {
		parentMerchant, err = s.merchantRepo.FindMerchantByID(ctx, merchant.ParentID.String)
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrInternal, c.ErrFindParentMerchant)
		}
		// When sub-merchant create payment directly
		if _, ok := ctx.Value(c.CtxParentMerchantId).(string); !ok {
			ctx = context.WithValue(ctx, c.CtxParentMerchantId, merchant.ParentID.String)
		}
	}

	if merchant.ParentID.Valid && merchant.KYCStatus.String != c.KYCStatusApproved {
		derivedMerchantId = merchant.ParentID.String
	}

	// Get Payment Method by type and bankCode
	paymentMethod, err := s.paymentMethodSvc.GetActivePaymentMethodDetailForPaymentRequest(ctx, paymentModel.GetActivePaymentMethodRequest{
		MerchantID: derivedMerchantId,
		Category:   constant.PAYMENT_METHOD_CATEGORY_PAYMENT,
		Type:       constant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Acquirer:   virtualAccountReq.Issuer,
	})
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if paymentMethod == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, c.ErrPaymentMethodNotFound)

	} else if paymentMethod.ChannelType == c.PaymentMethodChannelTypeDirect && len(util.ValueOfPtr(paymentRequest.SplitRoutingConfigurations)) > 0 {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, c.ErrDoNotApplySplitRouteInFacilitatorModel)
	}

	if err := s.ValidatePaymentExpiry(ctx, paymentModel.PaymentRequestExpiryValidation{
		MerchantID:    merchantID,
		Method:        c.UnifiedPaymentMethodVA,
		Request:       &paymentRequest,
		PaymentMethod: paymentMethod,
	}); err != nil {
		return nil, err
	}

	// find or create customer
	customer, err := s.findOrCreateCustomer(ctx, paymentRequest.Customer, merchantID)
	if err != nil {
		return nil, err
	}

	if paymentRequest.IsSnap && paymentRequest.ReferenceID != "" {
		payment, err := s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, merchantID, paymentRequest.ReferenceID)
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrDatabase, err)
		}
		if payment != nil {
			return nil, pkgErrors.New(response.HttpErrDupCheck, errors.New(c.DuplicatePartnerReferenceNoErrMsg))
		}
	}

	amount := paymentRequest.TotalAmount.Value // client input amount
	discount := decimal.NewFromFloat(0)
	feeInDecimal := decimal.NewFromFloat(0)
	totalAmount := amount.Add(feeInDecimal).Sub(discount)

	currency := paymentRequest.TotalAmount.Currency

	// Request VA to SnapCore
	var snapCoreBillDetailsRequest []snapCoreModel.BillDetail

	// Check BillDetails if exist then map to snap core request
	// we should ignore payment items if bill details is provided to follow snap standard
	if util.ValueOfPtr(paymentRequest.VirtualAccount).BillDetails != nil {
		for _, billItem := range *paymentRequest.VirtualAccount.BillDetails {
			snapCoreBillDetailsRequest = append(snapCoreBillDetailsRequest, snapCoreModel.BillDetail{
				BillNo:            billItem.BillNo,
				BillerReferenceId: billItem.BillerReferenceId,
				BillCode:          billItem.BillCode,
				BillName:          billItem.BillName,
				BillShortName:     billItem.BillShortName,
				BillDescription: snapCoreModel.Description{
					English:   util.ValueOfPtr(billItem.BillDescription).English,
					Indonesia: util.ValueOfPtr(billItem.BillDescription).Indonesia,
				},
				BillSubCompany: billItem.BillSubCompany,
				BillAmount: snapCoreModel.Amount{
					Value:    billItem.BillAmount.Value,
					Currency: billItem.BillAmount.Currency,
				},
				AdditionalInfo: &billItem.AdditionalInfo,
			})
		}
	} else if paymentRequest.PaymentItems != nil && len(*paymentRequest.PaymentItems) > 0 {
		for _, item := range *paymentRequest.PaymentItems {
			snapCoreBillDetailRequest := item.ToSnapCoreBillDetail()
			snapCoreBillDetailsRequest = append(snapCoreBillDetailsRequest, snapCoreBillDetailRequest)
		}
	}

	vaNumber := virtualAccountReq.VirtualAccountNumber
	if virtualAccountReq.VirtualAccountTrxType == constant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC {
		vaNumber = ""
	}

	var accountName string
	if virtualAccountReq.VirtualAccountName != "" {
		accountName = virtualAccountReq.VirtualAccountName
	}

	if virtualAccountReq.VirtualAccountTrxType == constant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC && virtualAccountReq.MinAmount != nil {
		minAmount = snapCoreModel.Amount{
			Currency: currency,
			Value:    virtualAccountReq.MinAmount.Value.StringFixed(2),
		}
		additionalInfo = &map[string]interface{}{
			"minAmount": minAmount,
		}
	}

	if virtualAccountReq.VirtualAccountTrxType == constant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC && virtualAccountReq.MaxAmount != nil {
		maxAmount = snapCoreModel.Amount{
			Currency: currency,
			Value:    virtualAccountReq.MaxAmount.Value.StringFixed(2),
		}
		(*additionalInfo)["maxAmount"] = maxAmount
	}

	paymentUUID := uuid.NewString()
	if paymentRequest.UUID != "" {
		paymentUUID = paymentRequest.UUID
	}

	// Passing Account Transaction ID to Processor
	if _, ok := ctx.Value(c.CtxChangePaymentMethod).(bool); ok {
		pendingLedger, err := s.orchestratorSvc.FindByReference(ctx, paymentUUID, c.TypePayment)
		if err != nil {
			return nil, err
		} else if pendingLedger == nil {
			return nil, pkgErrors.New(response.HttpStatusErrorUnprocessableContent, c.ErrPaymentNotFound)
		}

		(*additionalInfo)[c.ProcessorExternalIDKey] = pendingLedger.UUID
	}

	// Request Create VA
	snapCoreRequest := snapCoreModel.CreateVirtualAccountRequest{
		MID:      merchant.MID.String, // Get 4 digit MID from merchant table
		VaNumber: vaNumber,
		TotalAmount: snapCoreModel.Amount{
			Currency: currency,
			Value:    amount.StringFixed(2), // amount to be paid
		},
		Acquirer:       strings.ToLower(virtualAccountReq.Issuer),
		BillDetails:    snapCoreBillDetailsRequest,
		IsCloseAmount:  snapCoreModel.VaTrxType(virtualAccountReq.VirtualAccountTrxType).IsCloseAmount,
		IsSingleUse:    snapCoreModel.VaTrxType(virtualAccountReq.VirtualAccountTrxType).IsSingleUsed,
		ExpiredAt:      virtualAccountReq.ExpiredDate,
		AccountName:    accountName,
		AdditionalInfo: additionalInfo,
		MerchantID:     merchantID,
		CustomerNo:     merchantID,
	}
	if parentMerchant != nil && merchant.KYCStatus.String != c.KYCStatusApproved {
		snapCoreRequest.MID = parentMerchant.MID.String
		snapCoreRequest.MerchantID = parentMerchant.UUID
	}

	if customer != nil {
		if snapCoreRequest.AccountName == "" {
			snapCoreRequest.AccountName = customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName)
		}
		snapCoreRequest.AccountEmail = customer.Email
		snapCoreRequest.AccountPhone = customer.PhoneNumber
	}

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    paymentUUID,
		ReferenceId: merchantID,
		From:        serviceName,
	})

	snapCoreResp, err := s.snapCoreRepo.CreateVirtualAccount(ctx, snapCoreRequest)
	if err != nil {
		s.logger.Error(ctx, "error when do request create virtual account to snap core", logger.Error(err))
		return nil, err
	}

	if virtualAccountReq.MinAmount != nil {
		snapCoreResp.MinAmount = snapCoreModel.Amount{
			Currency: snapCoreResp.MinAmount.Currency,
			Value:    snapCoreResp.MinAmount.Value,
		}
	}

	if virtualAccountReq.MaxAmount != nil {
		snapCoreResp.MaxAmount = snapCoreModel.Amount{
			Currency: snapCoreResp.MaxAmount.Currency,
			Value:    snapCoreResp.MaxAmount.Value,
		}
	}

	// curently snap core not return bill details, so we need to map from request
	snapCoreResp.BillDetails = paymentRequest.VirtualAccount.BillDetails

	metadataMap := make(map[string]interface{})
	metadataMap["snapCore"] = snapCoreResp
	metadataMap["isSnap"] = paymentRequest.IsSnap
	if parentMerchantId, _ := ctx.Value(c.CtxParentMerchantId).(string); parentMerchantId != "" {
		metadataMap["onBehalf"] = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}
	metadataMap["clientRedirectUrl"] = paymentRequest.ClientRedirectUrl
	metadataMap[c.IsUnifiedPaymentKey] = paymentRequest.IsUnifiedPayment
	if paymentRequest.SplitRoutingConfigurations != nil && len(*paymentRequest.SplitRoutingConfigurations) > 0 {
		metadataMap[c.SplitRoutingPaymentConfigKey] = *paymentRequest.SplitRoutingConfigurations
	}

	accountTrxMetadata := orchestratorModel.MetadataPayment[orchestratorModel.MetadataPaymentMethodVA]{
		ReconReferenceNo: snapCoreResp.VirtualAccountNo,
		ExpiredAt:        snapCoreResp.ExpiredAt,
		MethodDetail: orchestratorModel.MetadataPaymentMethodVA{
			AccountName:    snapCoreResp.AccountName,
			Acquirer:       snapCoreResp.Acquirer,
			Status:         snapCoreResp.Status,
			CreatedAt:      snapCoreResp.CreatedAt,
			IsClosedAmount: snapCoreResp.IsClosedAmount,
			IsSingleUse:    snapCoreResp.IsSingleUse,
			AdditionalInfo: snapCoreResp.AdditionalInfo,
		},
	}
	rawAccountTrxMetadata, _ := json.Marshal(accountTrxMetadata)

	metaDataB, _ := json.Marshal(metadataMap)
	metaDataString := string(metaDataB)

	// Begin Tx
	ctx, err = s.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	var payment paymentModel.Payment
	paymentDTO := paymentModel.PaymentDTO{
		UUID:                     paymentUUID,
		ReferenceID:              &paymentRequest.ReferenceID,
		MerchantID:               merchantID,
		PaymentMethodID:          paymentMethod.UUID,
		ProcessorReferenceNumber: &snapCoreResp.VirtualAccountNo,
		Currency:                 currency,
		Amount:                   amount.Round(2),
		Fee:                      &feeInDecimal,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   paymentRequest.InitiateStatus,
		Metadata:                 &metaDataString,
		ExpiredAt:                virtualAccountReq.ExpiredDate,
		CreatedBy:                &paymentRequest.CreatedBy,
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
		PaymentURL:               paymentRequest.PaymentUrl,
	}

	if customer != nil {
		paymentDTO.CustomerID = customer.UUID
	}

	if _, ok := ctx.Value(c.CtxChangePaymentMethod).(bool); ok {
		// Update new payment method
		currentChannel, _ := ctx.Value(c.CtxCurrentPaymentMethod).(string)
		if err = s.paymentRepo.ChangePaymentMethod(ctx, &paymentDTO); err == nil && virtualAccountReq.VirtualAccountTrxType != constant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC {
			trxRequest := orchestratorModel.UpdateTransactionWithPendingStatus{
				Metadata:        rawAccountTrxMetadata,
				UpdatedAt:       time.Now().UTC(),
				Processor:       c.SnapCoreProcessor,
				ProcessorID:     snapCoreResp.ID,
				Channel:         c.ChannelVirtualAccount,
				SettlementModel: paymentMethod.ChannelType,
			}
			err = s.accountTransactionRepo.UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(
				ctx, paymentUUID, c.TypePayment, currentChannel, trxRequest,
			)
		}
	} else {
		// Create new payment
		if err = s.paymentRepo.CreatePayment(ctx, &paymentDTO); err == nil && virtualAccountReq.VirtualAccountTrxType != constant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC {
			// Create PENDING transaction with VA detail
			trxRequest := &orchestratorModel.CreateAccountTransactionRequest{
				UUID:                 uuid.New(),
				ReferenceID:          paymentUUID,
				Type:                 c.TypePayment,
				MerchantID:           util.ParseUUID(merchantID),
				Currency:             currency,
				Credit:               paymentRequest.TotalAmount.Value.Round(2).InexactFloat64(),
				Channel:              c.ChannelVirtualAccount,
				Status:               c.StatusPending,
				SettlementStatus:     util.ValueToPtr(c.StatusPending),
				TransactionTimestamp: paymentDTO.CreatedAt,
				Usecase:              c.TypePayment,
				Processor:            c.SnapCoreProcessor,
				ProcessorID:          snapCoreResp.ID,
				AdditionalInfo: types.NullJSONText{
					Valid: true, JSONText: rawAccountTrxMetadata,
				},
				SettlementModel: util.ValueToPtr(paymentMethod.ChannelType),
			}
			// Action For Post Transaction
			err = s.orchestratorSvc.PostAccountTransaction(ctx, trxRequest)
		}
	}
	if err != nil {
		s.logger.Error(ctx, "error when create payment", logger.Error(err))

		// Rollback Tx
		if errRollback := s.paymentRepo.RollbackTransaction(ctx); errRollback != nil {
			return nil, errRollback
		}

		return nil, err
	}

	payment.PaymentFromDTO(&paymentDTO)

	// Check BillDetails if exist then save to payment items
	paymentItems := paymentRequest.PaymentItems
	if paymentItems != nil && len(*paymentItems) > 0 {
		for _, item := range *paymentItems {
			var itemMetadata *string
			if item.Metadata != nil {
				itemMetadataInJson, _ := json.Marshal(item.Metadata)
				itemMetadataInString := string(itemMetadataInJson)
				itemMetadata = &itemMetadataInString
			}

			itemQty := item.Qty
			itemAmount := item.Amount
			itemTotalAmount := itemAmount.Value.Mul(decimal.NewFromInt(int64(itemQty)))

			var paymentItem paymentModel.PaymentItem
			paymentItemDTO := paymentModel.PaymentItemDTO{
				UUID:        uuid.NewString(),
				PaymentID:   paymentUUID,
				Name:        item.Name,
				Description: item.Description,
				Qty:         itemQty,
				Currency:    item.Amount.Currency,
				Amount:      itemAmount.Value,
				TotalAmount: itemTotalAmount,
				Metadata:    itemMetadata,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			}
			err = s.paymentRepo.CreatePaymentItem(ctx, &paymentItemDTO)
			if err != nil {
				s.logger.Error(ctx, "error when create payment item", logger.Error(err))

				// Rollback Tx
				if errRollback := s.paymentRepo.RollbackTransaction(ctx); errRollback != nil {
					return nil, errRollback
				}

				return nil, err
			}

			paymentItem.PaymentItemFromDTO(&paymentItemDTO)
			paymentItemsResponse = append(paymentItemsResponse, paymentModel.PaymentResponseItem{
				ItemID:      paymentItem.UUID,
				Name:        paymentItem.Name,
				Description: paymentItem.Description,
				Amount: paymentModel.Amount{
					Value:    itemAmount.Value,
					Currency: itemAmount.Currency,
				},
				Qty: itemQty,
			})
		}

		paymentResponse.PaymentItems = &paymentItemsResponse
	}

	// Commit Tx
	if err = s.paymentRepo.CommitTransaction(ctx); err != nil {
		return nil, err
	}

	paymentResponse.ToPaymentResponse(
		&paymentDTO,
		snapCoreResp,
		&paymentRequest,
		customer,
		merchantID,
	)

	paymentResponse.ToVirtualAccountResponse(
		&paymentRequest,
		snapCoreResp,
	)

	// Add payment simulation only on staging
	if s.config.Environment == c.EnvironmentStaging && paymentResponse.VirtualAccount != nil {
		paymentResponse.VirtualAccount.Metadata = map[string]any{
			c.PaymentSimulatorKey: fmt.Sprintf(
				s.config.MerchantPortalConfig.PaymentSimulationPatternURL,
				base64.StdEncoding.EncodeToString([]byte(payment.UUID)),
			),
		}
	}

	return &paymentResponse, nil
}

func (s *PaymentService) findOrCreateCustomer(ctx context.Context, payload paymentModel.PaymentRequestCustomer, merchantID string) (*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/findOrCreateCustomer")
	defer segment.End()

	if payload.Email == "" {
		return nil, nil
	}

	customer, err := s.customerRepo.FindCustomerByEmail(ctx, payload.Email)
	if err != nil {
		s.logger.Error(ctx, "error when get customer", logger.Error(err))
		return nil, err
	}

	if customer != nil {
		return customer, nil
	}

	firstName, lastName := customerModel.FullNameToFirstNameAndLastName(payload.Name)
	customer = &customerModel.Customer{
		UUID:        uuid.NewString(),
		MerchantID:  merchantID,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       payload.Email,
		PhoneNumber: payload.Phone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err = s.customerRepo.Create(ctx, customer)
	if err != nil {
		s.logger.Error(ctx, "error when insert customer", logger.Error(err))
		return nil, err
	}

	return customer, nil
}

func (s *PaymentService) getFlagMaximumAmount() decimal.Decimal {
	maxAmountFlag := ffcontext.NewEvaluationContext(s.config.Environment)
	maximumAmount, err := ffclient.IntVariation("backend-portal-va-payment-maximum-amount", maxAmountFlag, 0)
	if err != nil && maximumAmount == 0 {
		return decimal.Zero
	}

	return decimal.NewFromInt(int64(maximumAmount))
}
