package adjustment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

const (
	disbursementType = "DISBURSEMENT"
	normalAction     = "NORMAL"
	adjustmentAction = "ADJUSTMENT"
)

func (s *adjustmentService) CreateManualTopup(ctx context.Context, req *adjustModel.ManualTopupRequest) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/adjustment/CreateManualTopup")
	defer segment.End()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, req.MerchantID)
	if err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("merchant id not found"))
	}

	currentBalance, err := s.orchestrator.GetAvailableMerchantBalance(ctx, req.MerchantID, constant.AccountNameDisbursement)
	if err != nil {
		s.logger.Error(ctx, "Failed while getting available merchant balance", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	id = util.GenerateUUID().String()
	req.File.Filename = id + filepath.Ext(req.File.Filename)

	url, err := s.gcs.UploadProofOfTransfer(ctx, req.File, false)
	if err != nil {
		return "", pkgErrs.New(response.HttpErrThirdParty, err)
	}

	if ctx, err = s.repo.BeginTransaction(ctx); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.repo.RollbackTransaction(ctx); e != nil {
				id, err = "", pkgErrs.New(response.HttpErrDatabase, e)
			}
		}
	}()

	bytesBankAccount, _ := json.Marshal(&adjustModel.BankAccount{
		Name:      req.BankName,
		AccNumber: req.BankAccount,
	})
	data := &adjustModel.ManualAdjustmentHistory{
		UUID:            id,
		MerchantID:      req.MerchantID,
		TransactionDate: time.Now().UTC(),
		BankRefID:       req.BankRefID,
		BankAccount:     string(bytesBankAccount),
		Type:            disbursementType,
		Action:          normalAction,
		Currency:        req.Currency,
		Amount:          req.Amount,
		ProofOfTransfer: url,
		Notes:           req.Notes,
		CreatedBy:       req.CreatedBy,
		CreatedAt:       time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err = s.repo.CreateAdjustment(ctx, data); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	}

	ledger := &orchestraModel.CreateAccountTransactionRequest{
		UUID:                 util.GenerateUUID(),
		ReferenceID:          id,
		Currency:             req.Currency,
		Credit:               req.Amount,
		Type:                 constant.TypeManualAdjust,
		Channel:              constant.ChannelManualTransfer,
		Status:               constant.StatusSuccess,
		Remarks:              req.Notes,
		TransactionTimestamp: data.TransactionDate,
	}
	ledger.MerchantID, _ = uuid.Parse(req.MerchantID)

	if err = s.orchestrator.PostAccountTransaction(ctx, ledger); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	}

	if err = s.repo.CommitTransaction(ctx); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	slackMessage := &slackPb.PostWebhookCmd{
		URL:   s.cfg.ManualTopupNotifWebhookURL,
		Color: slackPb.Color_GOOD,
		Title: "Manual Topup Has Been Successfully Processed",
		Fields: []*slackPb.AttachmentField{
			{Title: "Merchant Name", Value: merchant.Name, Short: true},
			{Title: "Merchant ID", Value: merchant.UUID, Short: true},
			{Title: "Bank Reference ID", Value: req.BankRefID, Short: true},
			{Title: "Amount", Value: util.FormatRupiah(req.Amount), Short: true},
			{Title: "Currency", Value: req.Currency, Short: true},
			{Title: "Bank Acc Name", Value: req.BankName, Short: true},
			{Title: "Bank Acc Number", Value: req.BankAccount, Short: true},
			{Title: "Ops PIC Name", Value: req.CreatedBy, Short: true},
			{Title: "Transaction Time", Value: util.SnapCompatible(data.TransactionDate), Short: false},
		},
	}
	rawSlackMessage, _ := proto.Marshal(slackMessage)

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)

	if !req.SendCallback {
		return
	}

	// Send Merchant Callback
	callbackRequest := &merchantTopUp.MerchantTopUpCallbackRequest{
		UUID:         ledger.UUID.String(),
		MerchantID:   req.MerchantID,
		MerchantName: merchant.Name,
		AccountName:  constant.AccountNameDisbursement,
		Amount: common.Amount{
			Currency: req.Currency,
			Value:    fmt.Sprintf("%.2f", req.Amount),
		},
		BalanceBefore: common.Amount{
			Currency: req.Currency,
			Value:    fmt.Sprintf("%.2f", currentBalance),
		},
		BalanceAfter: common.Amount{
			Currency: req.Currency,
			Value:    fmt.Sprintf("%.2f", currentBalance+req.Amount),
		},
		PaymentMethod: merchantTopUp.MerchantTopUpCallbackPaymentMethodObject{
			Type: constant.ChannelManualTransfer,
		},
		PaymentMethodOptions: merchantTopUp.MerchantTopUpCallbackPaymentMethodOptionsObject{},
		TransactionTime:      data.TransactionDate,
		ParentMerchantID:     merchant.ParentID.String,
	}
	if errCallback := s.merchantTopUpCallback.SendCallback(ctx, constant.CallbackEventMerchantTopUpSuccess, callbackRequest); errCallback != nil {
		s.logger.Warn(ctx, "Failed to publish merchant callback delivery message", logger.Error(errCallback), logger.Any("request", callbackRequest))
	}
	return
}
