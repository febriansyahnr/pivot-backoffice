package merchant

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

// Create		godoc
// @Summary		Create merchant endpoint
// @Description	Create merchant endpoint
// @ID			crm-merchant-create
// @Tags		CRM - Merchant
// @Accept		json
// @Produce		json
// @Param		Request	body		merchant.CRMCreateMerchantRequest true "JSON Body for Create Merchant"
// @Success		200  	{object}	response.Response{data=merchant.MerchantResponse}
// @Failure		500  	{object}	response.Response
// @Router		/api/v1/merchants/create [post]
// @Security	Bearer
func (c *CRMMerchantController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/Create")
	defer segment.End()

	var (
		err  error
		user = "System"
	)

	var payload merchantModel.CRMCreateMerchantRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	merchantData := &merchantModel.Merchant{
		UUID:          uuid.New().String(),
		ExternalId:    util.GenerateULID(),
		Name:          payload.Name,
		ShortName:     payload.ShortName,
		Description:   payload.Description,
		Website:       payload.Website,
		Address:       payload.Address,
		DistrictId:    payload.DistrictId,
		PostCode:      payload.PostCode,
		Logo:          payload.Logo,
		MerchantEmail: payload.MerchantEmail,
		MerchantPhone: payload.MerchantPhone,
		PICEmail:      payload.PICEmail,
		PICPhone:      payload.PICPhone,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		ParentIndustry: sql.NullString{
			String: payload.ParentIndustry,
			Valid:  payload.ParentIndustry != "",
		},
		ChildIndustry: sql.NullString{
			String: payload.ChildIndustry,
			Valid:  payload.ChildIndustry != "",
		},
		MCC: sql.NullString{
			String: payload.MCC,
			Valid:  payload.MCC != "",
		},
		CountryOfEntity: sql.NullString{
			String: payload.CountryOfEntity,
			Valid:  payload.CountryOfEntity != "",
		},
		DigitalStatus: sql.NullString{
			String: payload.DigitalStatus,
			Valid:  payload.DigitalStatus != "",
		},
		RiskLevel: sql.NullString{
			String: payload.RiskLevel,
			Valid:  payload.RiskLevel != "",
		},
		BusinessType: sql.NullString{
			String: payload.BusinessType,
			Valid:  payload.BusinessType != "",
		},
		BusinessStructure: sql.NullString{
			String: payload.BusinessStructure,
			Valid:  payload.BusinessStructure != "",
		},
		BusinessCountry: sql.NullString{
			String: payload.BusinessCountry,
			Valid:  payload.BusinessCountry != "",
		},
		PICName: sql.NullString{
			String: payload.PICName,
			Valid:  payload.PICName != "",
		},
		PICJobTitle: sql.NullString{
			String: payload.PICJobTitle,
			Valid:  payload.PICJobTitle != "",
		},
	}

	transactionConfig := map[string]interface{}{
		"autoWithdrawal": constant.AutoWithdrawalStateOFF,
	}

	if c.config.WithdrawalConfig.AutoWithdrawalDefaultState == constant.AutoWithdrawalStateON {
		transactionConfig["autoWithdrawal"] = constant.AutoWithdrawalStateON
	}

	if payload.AutoWithdrawal != nil {
		transactionConfig["autoWithdrawal"] = *payload.AutoWithdrawal
	}

	if payload.KYCStatus == constant.MerchantKYCTypeNonKYC {
		merchantData.Status = constant.MerchantStatusActive
		merchantData.KYCStatus.Valid = true
		merchantData.KYCStatus.String = constant.KYCStatusApproved
	} else {
		merchantData.Status = constant.MerchantStatusCreated
		merchantData.KYCStatus.Valid = true
		merchantData.KYCStatus.String = constant.KYCStatusWaitingForDocument
	}

	merchantData.TransactionConfigs.Valid = true
	merchantData.TransactionConfigs.JSONText, _ = json.Marshal(transactionConfig)

	if err = c.merchantSvc.Create(ctx, merchantData, payload.UserID); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	if c.config.Environment == constant.EnvironmentStaging {
		err := c.merchantSvc.EnableAllPaymentMethod(ctx, merchantData)
		if err != nil {
			c.logger.Warn(ctx, "failed to enable all payment method", pdkLogger.Error(err))
		}
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		nil,
		&user,
		constant.TagMerchant,
		constant.ActivityUserCreateMerchant,
		payload,
	)

	response.SendGeneralResponseOK(w, merchantData.ToCRMMerchantResponse())
}
