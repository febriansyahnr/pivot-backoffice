package adjustment

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *adjustmentService) CreateBalanceAdjustmentFromManualTopUp(ctx context.Context, req *adjustModel.BalanceAdjustmentRequest) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/adjustment/CreateBalanceAdjustmentFromManualTopUp")
	defer segment.End()

	id = uuid.NewString()
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, req.MerchantID)
	if err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	// get related manual adjustment
	relatedAdjustment, err := s.repo.FindByID(ctx, req.AdjustmentID)
	if err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	} else if relatedAdjustment == nil {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("related adjustment ID not found"))
	}

	if errValidate := s.validateBalanceAdjustment(ctx, relatedAdjustment, req); errValidate != nil {
		return "", errValidate
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

	data := &adjustModel.ManualAdjustmentHistory{
		UUID:            id,
		MerchantID:      req.MerchantID,
		TransactionDate: time.Now().UTC(),
		BankRefID:       relatedAdjustment.BankRefID,
		BankAccount:     relatedAdjustment.BankAccount,
		Type:            disbursementType,
		Action:          adjustmentAction,
		Currency:        req.Currency,
		Amount:          req.Amount,
		Notes:           req.Notes,
		CreatedBy:       req.CreatedBy,
		ReferenceID:     relatedAdjustment.UUID,
		CreatedAt:       time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err = s.repo.CreateAdjustment(ctx, data); err != nil {
		return "", pkgErrs.New(response.HttpErrDatabase, err)
	}

	credit := 0.0
	debit := 0.0
	if req.Amount < 0 {
		debit = math.Abs(req.Amount)
	} else {
		credit = req.Amount
	}

	ledger := &orchestraModel.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		ReferenceID:          id,
		Currency:             req.Currency,
		Credit:               credit,
		Debit:                debit,
		Type:                 constant.TypeManualAdjust,
		Channel:              constant.ChannelBalanceAdjustment,
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

	adjustmentBankAccount := &adjustModel.BankAccount{}
	if err = json.Unmarshal([]byte(relatedAdjustment.BankAccount), &adjustmentBankAccount); err != nil {
		return "", pkgErrs.New(response.HttpErrRequest, err)
	}

	slackMessage := &slackPb.PostWebhookCmd{
		URL:   s.cfg.ManualTopupNotifWebhookURL,
		Color: slackPb.Color_GOOD,
		Title: fmt.Sprintf("Balance Adjustment for ID:%s Has Been Successfully Processed", relatedAdjustment.UUID),
		Fields: []*slackPb.AttachmentField{
			{Title: "Merchant Name", Value: merchant.Name, Short: true},
			{Title: "Merchant ID", Value: merchant.UUID, Short: true},
			{Title: "Bank Reference ID", Value: relatedAdjustment.BankRefID, Short: true},
			{Title: "Amount", Value: util.FormatRupiah(req.Amount), Short: true},
			{Title: "Currency", Value: req.Currency, Short: true},
			{Title: "Bank Acc Name", Value: adjustmentBankAccount.Name, Short: true},
			{Title: "Bank Acc Number", Value: adjustmentBankAccount.AccNumber, Short: true},
			{Title: "Ops PIC Name", Value: req.CreatedBy, Short: true},
			{Title: "Transaction Time", Value: util.SnapCompatible(data.TransactionDate), Short: false},
		},
	}
	rawSlackMessage, _ := proto.Marshal(slackMessage)

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)

	return id, nil
}

func (s *adjustmentService) validateBalanceAdjustment(ctx context.Context, relatedAdjustment *adjustModel.ManualAdjustmentHistory, req *adjustModel.BalanceAdjustmentRequest) error {
	// validateBalanceAdjustment relatedAdjustment
	if relatedAdjustment.Action != normalAction {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("can't perform to this related adjustment ID"))
	} else if relatedAdjustment.MerchantID != req.MerchantID {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("merchant is not valid"))
	}

	// calculated amount and referenced
	calculateAmount, err := s.repo.CalculateAmountBalanceAdjustmentForTopUp(ctx, req.AdjustmentID)
	if err != nil {
		return pkgErrs.New(response.HttpErrDatabase, err)
	} else if req.Amount < 0 && math.Abs(req.Amount) > calculateAmount {
		return pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("amount deduction exceeds topup amount"))
	}

	return nil
}
