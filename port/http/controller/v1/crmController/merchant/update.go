package merchant

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Update		godoc
// @Summary		Update merchant endpoint
// @Description	Update merchant endpoint. Supports both application/json and multipart/form-data. When using multipart/form-data, you can optionally include a 'logoFile' field to upload a new logo. Either 'logo' or 'logoFile' must be provided.
// @ID			crm-merchant-update
// @Tags		CRM - Merchant
// @Accept		json,multipart/form-data
// @Produce		json
// @Param		id	path		string	true	"Merchant ID"
// @Param		Request	body		merchant.MerchantRequest false "JSON Body for Update Merchant (when Content-Type is application/json)"
// @Param		name	formData	string	false	"Merchant name (when Content-Type is multipart/form-data)"
// @Param		shortName	formData	string	false	"Merchant short name (max 25 characters)"
// @Param		description	formData	string	false	"Description"
// @Param		address	formData	string	false	"Address"
// @Param		districtId	formData	integer	false	"District ID"
// @Param		postcode	formData	string	false	"Postcode"
// @Param		website	formData	string	false	"Website"
// @Param		logo	formData	string	false	"Logo URL (required if logoFile not provided)"
// @Param		logoFile	formData	file	false	"Logo file to upload (PNG, JPG, JPEG, max 5MB). Required if logo not provided"
// @Param		merchantEmail	formData	string	false	"Merchant email"
// @Param		merchantPhone	formData	string	false	"Merchant phone"
// @Param		merchantStatus	formData	string	false	"Merchant status"
// @Param		merchantReasonStatus	formData	string	false	"Reason status"
// @Param		businessType	formData	string	false	"Business type"
// @Param		businessStructure	formData	string	false	"Business structure"
// @Param		businessCountry	formData	string	false	"Business country"
// @Param		picName	formData	string	false	"PIC name"
// @Param		picEmail	formData	string	false	"PIC email"
// @Param		picPhone	formData	string	false	"PIC phone"
// @Param		picJobTitle	formData	string	false	"PIC job title"
// @Param		industryId	formData	string	false	"Industry ID"
// @Param		countryOfEntity	formData	string	false	"Country of entity"
// @Param		digitalStatus	formData	string	false	"Digital status"
// @Param		riskLevel	formData	string	false	"Risk level"
// @Param		kymNotes	formData	string	false	"KYM notes"
// @Success		200  	{object}	response.Response{data=merchant.MerchantResponse}
// @Failure		400  	{object}	response.Response
// @Failure		500  	{object}	response.Response
// @Router		/api/v1/merchants/{id} [put]
// @Security	Bearer
func (c *CRMMerchantController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/Update")
	defer segment.End()

	var (
		err  error
		user = "System"
	)

	merchantId := chi.URLParam(r, "id")
	if err := uuid.Validate(merchantId); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	var payload merchantModel.CRMUpdateMerchantRequest
	var logoFile *merchantModel.LogoFileUpload

	// Check content type to determine how to parse the request
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			c.logger.Error(ctx, "error parsing multipart form", logger.Error(err))
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}

		// Parse form values into payload
		payload = merchantModel.CRMUpdateMerchantRequest{
			Name:                 r.FormValue("name"),
			ShortName:            r.FormValue("shortName"),
			Description:          r.FormValue("description"),
			Address:              r.FormValue("address"),
			PostCode:             r.FormValue("postcode"),
			Website:              r.FormValue("website"),
			Logo:                 r.FormValue("logo"),
			MerchantEmail:        r.FormValue("merchantEmail"),
			MerchantPhone:        r.FormValue("merchantPhone"),
			MerchantStatus:       r.FormValue("merchantStatus"),
			MerchantReasonStatus: r.FormValue("merchantReasonStatus"),
			BusinessType:         r.FormValue("businessType"),
			BusinessStructure:    r.FormValue("businessStructure"),
			BusinessCountry:      r.FormValue("businessCountry"),
			PICName:              r.FormValue("picName"),
			PICEmail:             r.FormValue("picEmail"),
			PICPhone:             r.FormValue("picPhone"),
			PICJobTitle:          r.FormValue("picJobTitle"),
			IndustryID:           r.FormValue("industryId"),
			CountryOfEntity:      r.FormValue("countryOfEntity"),
			DigitalStatus:        r.FormValue("digitalStatus"),
			RiskLevel:            r.FormValue("riskLevel"),
			KYMNotes:             r.FormValue("kymNotes"),
		}

		// Parse districtId if provided
		if districtIdStr := r.FormValue("districtId"); districtIdStr != "" {
			districtId, err := strconv.ParseUint(districtIdStr, 10, 16)
			if err != nil {
				response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
				return
			}
			payload.DistrictId = uint16(districtId)
		}

		// Check if logo file is uploaded
		if _, fileHeader, err := r.FormFile("logoFile"); err == nil {
			logoFile = &merchantModel.LogoFileUpload{
				FileHeader: fileHeader,
			}
		}
	} else {
		// Parse JSON body (original behavior)
		if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	// Custom validation: Either logo URL or logoFile must be provided
	if payload.Logo == "" && logoFile == nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantLogoAndLogoFileRequired))
		return
	}

	updateRequest := &merchantModel.UpdateMerchantRequest{
		ID:                merchantId,
		Name:              payload.Name,
		ShortName:         payload.ShortName,
		Description:       payload.Description,
		Address:           payload.Address,
		DistrictId:        payload.DistrictId,
		Website:           payload.Website,
		PostCode:          payload.PostCode,
		Logo:              payload.Logo,
		LogoFile:          logoFile,
		MerchantEmail:     payload.MerchantEmail,
		MerchantPhone:     payload.MerchantPhone,
		Status:            payload.MerchantStatus,
		ReasonStatus:      payload.MerchantReasonStatus,
		PICName:           payload.PICName,
		PICEmail:          payload.PICEmail,
		PICPhone:          payload.PICPhone,
		PICJobTitle:       payload.PICJobTitle,
		BusinessType:      payload.BusinessType,
		BusinessStructure: payload.BusinessStructure,
		BusinessCountry:   payload.BusinessCountry,
		KYMNotes:          payload.KYMNotes,
		IndustryID:        payload.IndustryID,
		CountryOfEntity:   payload.CountryOfEntity,
		DigitalStatus:     payload.DigitalStatus,
		RiskLevel:         payload.RiskLevel,
	}

	merchant, err := c.merchantSvc.Update(ctx, updateRequest)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// publish activity, do nothing on error
	activityData := payload
	if logoFile != nil {
		// Add info about logo upload to activity
		type ActivityPayload struct {
			merchantModel.CRMUpdateMerchantRequest
			LogoUploaded bool   `json:"logoUploaded"`
			NewLogoURL   string `json:"newLogoUrl,omitempty"`
		}
		_ = c.rabbitMqExt.PublishActivity(
			ctx,
			nil,
			&user,
			constant.TagMerchant,
			constant.ActivityUserUpdateMerchant,
			ActivityPayload{
				CRMUpdateMerchantRequest: payload,
				LogoUploaded:             true,
				NewLogoURL:               merchant.Logo,
			},
		)
	} else {
		_ = c.rabbitMqExt.PublishActivity(
			ctx,
			nil,
			&user,
			constant.TagMerchant,
			constant.ActivityUserUpdateMerchant,
			activityData,
		)
	}

	response.SendGeneralResponseOK(w, merchant.ToCRMMerchantResponse())
}
