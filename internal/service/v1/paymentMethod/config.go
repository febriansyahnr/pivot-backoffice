package paymentMethodService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	snapCoreVaModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentMethodService) SetupConfig(ctx context.Context, request *paymentMethodModel.SetupPaymentMethodConfigRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/SetupConfig")
	defer segment.End()

	// Find merchant
	merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return err
	} else if merchant == nil {
		s.logger.Warn(ctx, "[SetupConfig] FindMerchantByID not found", logger.String("merchantID", request.MerchantID))
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrMerchantNotFound)
	} else if merchant.ParentID.Valid && merchant.KYCStatus.String != c.KYCStatusApproved {
		s.logger.Warn(ctx, "[SetupConfig] Merchant is not parent and non KYC")
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrMerchantShouldKYC)
	}

	// Find PaymentMethod by PaymentMethodID
	paymentMethod, err := s.paymentMethodRepo.FindPaymentMethodByIdAndMerchant(ctx, request.PaymentMethodID, request.MerchantID)
	if err != nil {
		return pkgErr.New(responseHttp.HttpErrDatabase, err)
	} else if paymentMethod == nil {
		s.logger.Warn(ctx, "[SetupConfig] PaymentMethod not found")
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrPaymentMethodNotFound)
	}

	if request.PartnerConfig == nil {
		request.PartnerConfig = &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{}
	}
	if request.ChannelConfig == nil {
		request.ChannelConfig = &paymentMethodModel.SetupPaymentMethodChannelConfigRequest{}
	}

	// Read Partner Config by Payment Method Type
	needToStorePartnerConfig := true
	switch paymentMethod.Type {
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		partnerConfigVARequest := &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{}
		if request.PartnerConfig.VirtualAccount != nil {
			partnerConfigVARequest = request.PartnerConfig.VirtualAccount
		}
		if request.ChannelType == "" {
			request.ChannelType = paymentMethod.ChannelType
		}

		request.PartnerConfig.VirtualAccount = partnerConfigVARequest
		request.PartnerConfig.VirtualAccount.MerchantID = merchant.UUID
		request.PartnerConfig.VirtualAccount.MerchantMID = merchant.MID.String
		request.PartnerConfig.VirtualAccount.ChannelType = request.ChannelType
		request.PartnerConfig.VirtualAccount.Acquirer = strings.ToUpper(paymentMethod.Acquirer)

		if errSetupPartnerConfig := s.setupPartnerConfigForVA(ctx, paymentMethod.Processor, partnerConfigVARequest); errSetupPartnerConfig != nil {
			// Just log the error
			s.logger.Warn(ctx, "[SetupConfig] Failed to setup partner config VA]", logger.Error(err))
			return errSetupPartnerConfig
		}
	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD, paymentConstant.PAYMENT_METHOD_VIRTUAL_TERMINAL:
		partnerConfigCardRequest := &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{}
		if request.PartnerConfig.Card != nil {
			partnerConfigCardRequest = request.PartnerConfig.Card
		}

		partnerConfigCardRequest.MerchantID = merchant.UUID
		partnerConfigCardRequest.MerchantShortName = merchant.ShortName
		partnerConfigCardRequest.ChannelType = request.ChannelType
		partnerConfigCardRequest.PaymentMethodType = paymentMethod.Type
		partnerConfigCardRequest.SplitCardPaymentConfig = request.SplitCardPaymentConfig
		partnerConfigCardRequest.CardFundedPayoutConfig = request.CardFundedPayoutConfig
		if errSetupPartnerConfig := s.setupPartnerConfigForCard(ctx, paymentMethod.Processor, partnerConfigCardRequest); errSetupPartnerConfig != nil {
			// Just log the error
			s.logger.Warn(ctx, "[SetupConfig] Failed to setup partner config card]", logger.Error(err))
			return errSetupPartnerConfig
		}

	case paymentConstant.PAYMENT_METHOD_EWALLET:
		// Validate EWallet Partner Config
		switch paymentMethod.Name {
		case paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_SHOPEEPAY:
			if request.PartnerConfig.EWallet.ExternalMerchantID == "" {
				return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("externalMerchantId is required for ShopeePay"))

			} else if request.PartnerConfig.EWallet.ExternalStoreID == "" {
				return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("externalStoreId is required for ShopeePay"))
			}
		case paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_DANA:
			if request.PartnerConfig.EWallet.SubMerchantID == "" {
				return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("subMerchantId is required for Dana"))
			}
		}

	case paymentConstant.PAYMENT_METHOD_QRIS:
		var (
			acquirerTerminalId string
			acquirerMerchantId string
			acquirerStoreIds   []string
		)
		partnerConfigQrisRequest := &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{}
		if request.PartnerConfig.Qris != nil {
			partnerConfigQrisRequest = request.PartnerConfig.Qris
		}

		partnerConfigQrisRequest.ChannelType = request.ChannelType
		partnerConfigQrisRequest.MerchantExternalID = merchant.ExternalId
		partnerConfigQrisRequest.MerchantID = merchant.UUID
		partnerConfigQrisRequest.Acquirer = strings.ToUpper(paymentMethod.Acquirer)

		// set qr type as dynamic when its not defined
		if partnerConfigQrisRequest.QRType == "" {
			partnerConfigQrisRequest.QRType = c.QrTypeDynamic
		}

		// Validate required fields
		if err := s.validateQrisPartnerConfig(ctx, request, partnerConfigQrisRequest, paymentMethod); err != nil {
			return err
		}

		partnerConfigQrisRequest.MerchantType = request.PartnerConfig.Qris.MerchantType
		partnerConfigQrisRequest.CreatedBy = request.PartnerConfig.Qris.CreatedBy

		qrisRegistration, errSetupPartnerConfig := s.setupPartnerConfigForQris(ctx, partnerConfigQrisRequest)
		if errSetupPartnerConfig != nil {
			s.logger.Warn(ctx, "[SetupConfig] Failed to setup partner config qris]", logger.Error(errSetupPartnerConfig))
			return errSetupPartnerConfig
		} else {
			if qrisRegistration.AcquirerTerminalId != nil && *qrisRegistration.AcquirerTerminalId != "" {
				acquirerTerminalId = *qrisRegistration.AcquirerTerminalId
			}
			if qrisRegistration.AcquirerMerchantId != nil && *qrisRegistration.AcquirerMerchantId != "" {
				acquirerMerchantId = *qrisRegistration.AcquirerMerchantId
			}
		}

		request.PartnerConfig.Qris.AcquirerMerchantID = acquirerMerchantId

		// Update partner config with acquirer data
		if paymentMethod.Acquirer != paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC || partnerConfigQrisRequest.QRType == c.QrTypeDynamic {
			request.PartnerConfig.Qris.AcquirerMerchantID = acquirerMerchantId
			request.PartnerConfig.Qris.AcquirerTerminalID = acquirerTerminalId
		}

		if partnerConfigQrisRequest.QRType == c.QrTypeStatic {
			cfg := util.ValueOfPtr(paymentMethod.MerchantConfigObj)
			partnerCfg := util.ValueOfPtr(cfg.PartnerConfig)
			QRCfg := util.ValueOfPtr(partnerCfg.Qris)
			acquirerStoreIds = QRCfg.AcquirerStoreIDs
		}

		for _, item := range partnerConfigQrisRequest.AcquirerStoreIDs {
			if util.InArray(acquirerStoreIds, item) {
				continue
			}

			acquirerStoreIds = append(acquirerStoreIds, item)
		}

		if len(acquirerStoreIds) > 0 {
			request.PartnerConfig.Qris.AcquirerStoreIDs = acquirerStoreIds
		}

		if qrisRegistration != nil {
			// set to sync to snap_core_processor/qr-registration-sync
			err := s.syncQrisRegistration(ctx, qrisRegistration, snapCoreModel.SyncRegistrationOption{
				MerchantID: paymentMethod.MerchantID,
				StoreIDs:   acquirerStoreIds,
				IsActive:   &paymentMethod.IsActive,
			})
			if err != nil {
				// Just log the error
				s.logger.Error(ctx, "[SetupConfig] Failed to sync QRIS registration to snap_core_processor/qr-registration-sync", logger.Error(err))
			}
		}

	case paymentConstant.PAYMENT_METHOD_INSTALLMENT:

		if paymentMethod.Subtype != constant.InstallmentPlanPaymentMethodCard {
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("payment method subtype is not card"))
		}
		paymentMethodMerchantList, err := s.paymentMethodRepo.GetListPaymentMethodMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
			MerchantID: merchant.UUID,
			Type:       paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			Status:     constant.PaymentMethodGeneralStatusActive,
		})
		if err != nil {
			return pkgErr.New(responseHttp.HttpErrDatabase, c.ErrInstallmentGetDependentPaymentMethod)
		}
		if len(paymentMethodMerchantList) < 1 {
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrDependentCardPaymentMethodNotActive)
		}

		if request.PartnerConfig.Installment == nil ||
			(request.PartnerConfig.Installment != nil && len(request.PartnerConfig.Installment.InstallmentPlanIDs) == 0) {
			return pkgErr.New(responseHttp.HttpErrRequest, fmt.Errorf("installment plan ids is required"))
		}
		installmentPlans, _, err := s.installmentPlanSvc.List(ctx, &installmentPlanModel.FilterRequest{
			InstallmentIDs: request.PartnerConfig.Installment.InstallmentPlanIDs,
			Acquirer:       paymentMethod.Acquirer,
			PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
			Status:         constant.InstallmentPlanStatusActive,
		})
		if err != nil {
			s.logger.Error(ctx, "[SetupConfig] Failed to list installment plans")
			return err
		}
		if len(installmentPlans) == 0 {
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrInstallmentPlanNotFound)
		}
		installmentIds := []string{}
		for _, installmentPlan := range installmentPlans {
			if installmentPlan.MerchantID != "" && installmentPlan.MerchantID != request.MerchantID {
				continue
			}
			installmentIds = append(installmentIds, installmentPlan.UUID)
		}

		invalidInstallmentPlanIds := []string{}
		for _, paymentMethodPlanId := range request.PartnerConfig.Installment.InstallmentPlanIDs {
			if !slices.Contains(installmentIds, paymentMethodPlanId) {
				invalidInstallmentPlanIds = append(invalidInstallmentPlanIds, paymentMethodPlanId)
			}
		}
		if len(invalidInstallmentPlanIds) > 0 {
			return pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("invalid installment plan ids: %v", strings.Join(invalidInstallmentPlanIds, ", ")))
		}
		needToStorePartnerConfig = true
	}

	// Update Payment Method Merchant Config, set Channel OR Partner Config
	if needToStorePartnerConfig {
		paymentMethod.MerchantConfigObj.PartnerConfig = request.PartnerConfig
		paymentMethod.MerchantConfigObj.SplitCardPaymentConfig = request.SplitCardPaymentConfig
		paymentMethod.MerchantConfigObj.CardFundedPayoutConfig = request.CardFundedPayoutConfig
		paymentMethod.MerchantConfigObj.VirtualTerminalConfig = request.VirtualTerminalConfig
	}

	paymentMethod.MerchantConfigObj.ChannelConfig = request.ChannelConfig
	paymentMethod.ChannelType = request.ChannelType
	paymentMethod.MerchantConfig.Valid = true
	paymentMethod.MerchantConfig.JSONText, _ = json.Marshal(paymentMethod.MerchantConfigObj)
	if errUpsert := s.paymentMethodRepo.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, paymentMethod); errUpsert != nil {
		return errUpsert
	}

	return nil
}

func (s *PaymentMethodService) setupPartnerConfigForVA(ctx context.Context, processor string, config *paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/setupPartnerConfigForVA")
	defer segment.End()

	if processor != c.SnapCoreProcessor {
		s.logger.Info(ctx, "[setupPartnerConfigForVA] processor is not snap core")
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrProcessorNotRegistered)
	}

	configDetail := []*snapCoreVaModel.UpdateVirtualAccountConfigPrefixDetail{}

	// Call processor create VA config
	for _, item := range config.Items {
		vaConfigFromSnapCore, err := s.snapCoreRepo.GetVirtualAccountConfig(ctx, &snapCoreVaModel.GetVirtualAccountConfigRequest{
			MerchantID:      config.MerchantID,
			MID:             config.MerchantMID,
			BinPrefix:       item.BINPrefix,
			Type:            item.Type,
			IntegrationType: config.ChannelType,
			Acquirer:        config.Acquirer,
		})
		if err != nil {
			return err
		}

		if len(vaConfigFromSnapCore) == 0 {
			createVAConfigRequest := &snapCoreVaModel.CreateVirtualAccountConfigRequest{
				MerchantID:      config.MerchantID,
				MID:             config.MerchantMID,
				BinPrefix:       item.BINPrefix,
				Type:            item.Type,
				IntegrationType: config.ChannelType,
				Acquirer:        config.Acquirer,
			}
			_, err := s.snapCoreRepo.CreateVirtualAccountConfig(ctx, createVAConfigRequest)
			if err != nil {
				return err
			}
		}

		configDetail = append(configDetail, &snapCoreVaModel.UpdateVirtualAccountConfigPrefixDetail{
			Type:       item.Type,
			StartRange: item.StartRange,
			EndRange:   item.EndRange,
		})
	}

	err := s.snapCoreRepo.UpdateVirtualAccountConfigPrefix(ctx, &snapCoreVaModel.UpdateVirtualAccountConfigPrefixRequest{
		MerchantID:      config.MerchantID,
		Acquirer:        config.Acquirer,
		IntegrationType: config.ChannelType,
		Detail:          configDetail,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *PaymentMethodService) setupPartnerConfigForCard(ctx context.Context, processor string, config *paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/setupPartnerConfigForCard")
	defer segment.End()

	if processor != c.CreditCardCoreProcessor {
		s.logger.Info(ctx, "[setupPartnerConfigForCard] processor is not creditcard core")
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrProcessorNotRegistered)
	}

	validations := []validateFn{
		func() error { return validateSplitCardPaymentConfig(config) },
		func() error { return validateCardFundedPayoutConfig(config) },
		func() error {
			return validateCardTransactionMIDPairConfig("recurring payment", config.Items, func(c cardConfig) string { return c.PartnerProcessor + "-" + c.RecurringType })
		},
		func() error {
			return validateCardTransactionMIDPairConfig("card-funded payout", config.Items, func(c cardConfig) string { return c.PartnerProcessor + "-" + c.CardFundedPayoutType })
		},
		func() error {
			return validateCardTransactionMIDPairConfig("split payment", config.Items, func(c cardConfig) string { return c.PartnerProcessor + "-" + c.SplitPaymentType })
		},
	}
	for _, validation := range validations {
		if err := validation(); err != nil {
			return err
		}
	}

	dupChecker := &uniqueData{travelAgents: map[string]bool{}}

	for i, item := range config.Items {

		item = setupValueForSpecificPaymentType(config.PaymentMethodType, &config.Items[i])

		if err := validateCardPartnerConfig(config.PaymentMethodType, item, dupChecker, s.config.VccTerminal.TravelAgents); err != nil {
			return pkgErr.New(response.HttpErrRequest, err)
		}
		channelType := c.ChannelTypeToMidType(config.ChannelType)

		// Find MID By Acquirer MID
		mid, err := s.creditCardRepo.GetMIDByAcquirerMID(ctx, item.AcquirerMerchantID)
		if err != nil && !strings.Contains(err.Error(), "NOT_FOUND") {
			return err
		}

		if item.ChannelType != "" {
			channelType = c.ChannelTypeToMidType(item.ChannelType)
		}

		if item.SupportedUseCase != nil && len(item.SupportedUseCase.AllowedECICodes) > 1 {
			slices.Sort(item.SupportedUseCase.AllowedECICodes)
		}

		if mid != nil && item.ChannelType == "" {
			config.Items[i].ChannelType = c.MidTypeToChannelType(mid.Type)
			channelType = mid.Type
		}

		// If exist than update, else create
		if mid == nil {
			createMidRequest := &creditcardCoreProcessorModel.CreateMIDRequest{
				Mid:                item.AcquirerMerchantID,
				Name:               strings.ToUpper(fmt.Sprintf("%s-%s-%s", channelType, config.MerchantShortName, item.PartnerProcessor)),
				Description:        item.Description,
				Type:               channelType,
				TransactionType:    constant.CreditCardMidTransactionTypeDirectPay,
				InstallmentType:    "",
				InstallmentTenor:   0,
				Processor:          item.PartnerProcessor,
				PrincipalAvailable: item.PrincipalAvailable,
				IsActive:           true,
				IsDefault:          false,
				BaseURL:            item.PartnerBaseURL,
				Password:           "predefined",
			}
			config.Items[i].ChannelType = config.ChannelType

			createMidResponse, err := s.creditCardRepo.CreateMID(ctx, createMidRequest)
			if err != nil {
				return err
			}

			if createMidResponse.Created {
				if mid, err = s.creditCardRepo.GetMIDByAcquirerMID(ctx, item.AcquirerMerchantID); err != nil {
					return err
				}
			}
		}

		if mid == nil {
			s.logger.Warn(ctx, "[setupPartnerConfigForCard] mid for assign is nil")
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrDataNotFound)
		}

		if mid.Type != channelType {
			s.logger.Info(ctx, "invalid mid type setup", logger.String("currentMIDType", mid.Type), logger.String("expectedMIDType", channelType))
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrPaymentMethodInvalidChannelType)
		}

		midMap, err := s.creditCardRepo.FindMIDMapByMerchant(ctx, &creditcardCoreProcessorModel.FindMIDMapByMerchantRequest{
			MerchantID: config.MerchantID,
			MidID:      mid.Uuid.String(),
		})

		if err != nil && !strings.Contains(err.Error(), "NOT_FOUND") {
			return err
		}

		if midMap != nil {
			_, err = s.creditCardRepo.UpdateMIDMapPriority(ctx, creditcardCoreProcessorModel.UpdateMIDMapPriorityRequest{
				MidMapID: midMap.Uuid,
				IsActive: item.IsActive,
				Priority: item.Priority,
			})

			if err != nil {
				return err
			}

			continue
		}

		// Assign to mid_mappings
		createMidMapRequest := &creditcardCoreProcessorModel.CreateMIDMapRequest{
			MidID:    mid.Uuid,
			Mid:      item.AcquirerMerchantID,
			IsActive: item.IsActive,
			Priority: item.Priority,
		}
		createMidMapRequest.MerchantID, _ = uuid.Parse(config.MerchantID)
		if _, err = s.creditCardRepo.CreateMIDMap(ctx, createMidMapRequest); err != nil {
			return err
		}
	}

	return nil
}

func (s *PaymentMethodService) setupPartnerConfigForQris(ctx context.Context, config *paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest) (qrisRegistration *qris.Registration, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/setupPartnerConfigForQris")
	defer segment.End()

	qrisReg, err := s.qrisSvc.FindQrRegistrationByExternalIDAndAcquirer(ctx, config.MerchantExternalID, config.Acquirer)
	if err != nil && !strings.Contains(err.Error(), "NOT_FOUND") {
		s.logger.Error(ctx, "[setupPartnerConfigForQris] Failed to find QRIS registration", logger.Error(err))
		return nil, err
	}

	if qrisReg == nil {
		id, err := s.qrisSvc.CreateManualRegistration(ctx, &qris.RegistrationReq{
			MerchantId:   config.MerchantID,
			Acquirer:     config.Acquirer,
			MerchantType: config.MerchantType,
			CreatedBy:    config.CreatedBy,
		})

		if err != nil {
			s.logger.Error(ctx, "[registrationForQrisConfig] Failed to init registration", logger.Error(err))
			return nil, err
		}

		s.logger.Info(ctx, "[setupPartnerConfigForQris] Successfully created manual registration", logger.String("registrationId", id))
		// Update qris registration with acquirer data
		err = s.qrisSvc.UpdateQrRegistration(ctx, id, config.AcquirerMerchantID, config.AcquirerTerminalID)
		if err != nil {
			s.logger.Error(ctx, "[setupPartnerConfigForQris] Failed to update QRIS registration", logger.Error(err))
			return nil, err
		}
		// Fetch updated registration
		updatedQrisReg, err := s.qrisSvc.FindQrRegistrationByExternalIDAndAcquirer(ctx, config.MerchantExternalID, config.Acquirer)
		if err != nil {
			s.logger.Error(ctx, "[setupPartnerConfigForQris] Failed to fetch updated QRIS registration", logger.Error(err))
			return nil, err
		}

		s.logger.Info(ctx, "[setupPartnerConfigForQris] Successfully updated QRIS registration",
			logger.String("registrationId", id),
			logger.String("acquirerMerchantId", config.AcquirerMerchantID),
			logger.String("acquirerTerminalId", config.AcquirerTerminalID))
		return updatedQrisReg, nil

	}

	if config.QRType == c.QrTypeDynamic && (util.ValueOfPtr(qrisReg.AcquirerMerchantId) != config.AcquirerMerchantID ||
		util.ValueOfPtr(qrisReg.AcquirerTerminalId) != config.AcquirerTerminalID) {
		s.logger.Info(ctx, "[setupPartnerConfigForQris] qr registration found", logger.String("id", qrisReg.Id))

		err = s.qrisSvc.UpdateQrRegistration(ctx, qrisReg.Id, config.AcquirerMerchantID, config.AcquirerTerminalID)
		if err != nil {
			s.logger.Error(ctx, "[setupPartnerConfigForQris] Failed to update QRIS registration", logger.Error(err))
			return nil, err
		}

		qrisReg.AcquirerMerchantId = &config.AcquirerMerchantID
		qrisReg.AcquirerTerminalId = &config.AcquirerTerminalID
	}

	s.logger.Info(ctx, "[setupPartnerConfigForQris] QRIS registration found",
		logger.String("registrationId", qrisReg.Id),
		logger.String("acquirerMerchantId", util.ValueOfPtr(qrisReg.AcquirerMerchantId)),
		logger.String("acquirerTerminalId", util.ValueOfPtr(qrisReg.AcquirerTerminalId)))

	return qrisReg, nil
}

func (s *PaymentMethodService) syncQrisRegistration(ctx context.Context, qrRegistration *qris.Registration, opt snapCoreModel.SyncRegistrationOption) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/syncQrisRegistration")
	defer segment.End()

	merchantInfo, err := s.merchantRepo.FindMerchantForQrRegistration(ctx, opt.MerchantID, qrRegistration.Acquirer)
	if err != nil {
		return err
	}

	if qrRegistration == nil {
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrDataNotFound)
	}

	if qrRegistration.AcquirerMerchantId == nil {
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrDataNotFound)
	}

	return s.snapCoreRepo.QrSyncRegistration(ctx, &snapCoreModel.SyncRegistrationDataRequest{
		Acquirer:                 qrRegistration.Acquirer,
		AcquirerMerchantID:       util.ValueOfPtr(qrRegistration.AcquirerMerchantId),
		AcquirerParentMerchantID: qrRegistration.AcquirerParentMerchantId,
		ApplymentCode:            qrRegistration.CallbackDetail.ApplymentCode,
		MerchantType:             qrRegistration.MerchantType,
		MerchantID:               opt.MerchantID,
		MCC:                      merchantInfo.MCC,
		BusinessShortname:        merchantInfo.ShortName,
		ExternalID:               qrRegistration.ExternalId,
		RegistrationID:           qrRegistration.Id,
		TerminalID:               util.ValueOfPtr(qrRegistration.AcquirerTerminalId),
		StoreIDs:                 opt.StoreIDs,
		Status:                   qrRegistration.Status,
		IsActive:                 opt.IsActive,
	})
}

func (s *PaymentMethodService) validateQrisPartnerConfig(ctx context.Context, request *paymentMethodModel.SetupPaymentMethodConfigRequest, partnerConfigQrisRequest *paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest, paymentMethod *paymentModel.PaymentMethodWithPivot) error {
	if partnerConfigQrisRequest.Acquirer != paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC {
		if request.PartnerConfig.Qris.AcquirerMerchantID == "" {
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("acquirerMerchantId is required"))
		}
		// if request.PartnerConfig.Qris.AcquirerTerminalID == "" {
		// 	return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("acquirerTerminalId is required"))
		// }
		if partnerConfigQrisRequest.MerchantType == "" {
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("merchantType is required"))
		}
		if partnerConfigQrisRequest.CreatedBy == "" {
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("createdBy is required"))
		}

		if partnerConfigQrisRequest.QRType == c.QrTypeStatic && len(partnerConfigQrisRequest.AcquirerStoreIDs) > 0 {
			s.logger.Error(ctx, "non bnc qris setup for storeID", logger.Error(c.ErrQrisInvalidPartnerConfigRequest))
			return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrQrisInvalidPartnerConfigRequest)
		}

		return nil
	}

	if paymentMethod.ActivationMethod == c.PaymentMethodActivationMethodApi {
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("payment activation method API is not allowed"))
	}

	if partnerConfigQrisRequest.QRType == c.QrTypeStatic && len(partnerConfigQrisRequest.AcquirerStoreIDs) == 0 {
		s.logger.Error(ctx, "bnc qris setup static with empty storeID", logger.Error(c.ErrQrisInvalidPartnerConfigRequest))
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrQrisInvalidPartnerConfigRequest)
	}

	if partnerConfigQrisRequest.QRType == c.QrTypeDynamic && partnerConfigQrisRequest.AcquirerMerchantID == "" {
		s.logger.Error(ctx, "bnc qris setup static with empty AcquirerMerchantID", logger.Error(c.ErrQrisInvalidPartnerConfigRequest))
		return pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrQrisInvalidPartnerConfigRequest)
	}

	return nil
}
