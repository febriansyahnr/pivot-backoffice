package merchant

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	industryUtil "github.com/paper-indonesia/pivot-backoffice/pkg/util/industry"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func (s *MerchantService) Update(ctx context.Context, request *merchantModel.UpdateMerchantRequest) (*merchantModel.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/Update")
	defer segment.End()

	// check if merchant is exist
	merchant, err := s.repo.FindMerchantByID(ctx, request.ID)
	if err != nil {
		s.logger.Error(ctx, "error when find merchant by id", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrInternal, err)
	} else if merchant == nil {
		return nil, pkgErrors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	// Handle logo file upload if provided
	if request.LogoFile != nil && request.LogoFile.FileHeader != nil {
		logoURL, err := s.UploadMerchantLogo(ctx, request.ID, request.LogoFile.FileHeader)
		if err != nil {
			s.logger.Error(ctx, "error when upload merchant logo", logger.Error(err))
			return nil, err
		}
		// Override the logo field with the uploaded file URL
		request.Logo = logoURL
		s.logger.Info(ctx, "merchant logo uploaded successfully during update",
			logger.String("merchantID", request.ID),
			logger.String("logoURL", logoURL))
	}

	if request.DistrictId > 0 {
		loc, err := s.locationRepo.GetDistrictById(ctx, request.DistrictId)
		if err != nil {
			s.logger.Error(ctx, "error when get district by id", logger.Error(err))
			return nil, pkgErrors.New(responseHttp.HttpErrInternal, err)
		}
		if loc == nil {
			return nil, pkgErrors.New(responseHttp.HttpErrUnprocessableContent, constant.ErrDistrictNotFound)
		}
	}
	if request.IndustryID != "" {
		industry, err := s.industrySvc.GetIndustryByID(ctx, request.IndustryID)
		if err != nil {
			return nil, err
		}
		if industry == nil {
			return nil, pkgErrors.New(responseHttp.HttpErrUnprocessableContent, constant.ErrIndustryNotFound)
		}
		request.ParentIndustry = industry.ParentIndustry
		request.ChildIndustry = industry.ChildIndustry
		request.MCC = industry.CommonMCC
	}
	if request.DigitalStatus != "" && !industryUtil.IsValidDigitalStatus(request.DigitalStatus) {
		return nil, pkgErrors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("digital status must be 'Digital' or 'Non-digital'"))
	}
	if request.CountryOfEntity != "" {
		country, err := s.countrySvc.FindByCode(ctx, request.CountryOfEntity)
		if err != nil {
			return nil, err
		}
		if country == nil {
			return nil, pkgErrors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid country code for country of entity"))
		}
		request.CountryOfEntity = country.Code
	}
	if request.RiskLevel != "" && !constant.IsValidRiskLevel(request.RiskLevel) {
		return nil, pkgErrors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid risk level: must be one of %v", constant.ValidMerchantRiskLevels))
	}

	err = merchant.UpdateMerchant(request)
	if err != nil {
		s.logger.Error(ctx, "error when update merchant detail", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrUnprocessableContent, constant.ErrUpdateMerchant)
	}
	if errUpdate := s.repo.Update(ctx, merchant); errUpdate != nil {
		s.logger.Error(ctx, "error when update merchant", logger.Error(errUpdate))
		return nil, pkgErrors.New(responseHttp.HttpErrInternal, errUpdate)
	}

	// update merchant status cache
	cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchant.UUID)
	_ = s.redis.Del(ctx, cacheKey)

	// publish event
	eventRequest := &pb.EventMerchantActionRequest{
		Event: constant.EventMerchantUpdated,
		Data:  merchant.ToProtoDataEvent(),
	}
	payload, err := proto.Marshal(eventRequest)
	pkgErrors.LogProtoMarshalError(ctx, s.logger, err, eventRequest)

	s.logger.Info(ctx, "Send event to sync update merchant", logger.String("data", base64.StdEncoding.EncodeToString(payload)))

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.MerchantActionRoutingKey, nil, payload)

	return merchant, nil
}
