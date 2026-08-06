package accountinquiry

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	requestAccountAnquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *AccountInquiryService) FindLatestByInquiryID(ctx context.Context, inquiryID, merchantID string) (*requestAccountAnquiry.RequestAccountInquiryWithMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/requestAccountInquiry/FindLatestByInquiryID")
	defer segment.End()

	var (
		inquiryAccount *requestAccountAnquiry.RequestAccountInquiryWithMaster
		err            error
	)

	inquiryAccount, err = s.repo.FindLatestByInquiryID(ctx, inquiryID, merchantID)

	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if inquiryAccount == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrInquiryIdNotFound)
	}

	// Force to use processor response
	if inquiryAccount.Metadata.Valid && inquiryAccount.MetadataObj.SnapCoreResponse != nil {
		inquiryAccount.MasterBeneficiaryAccountName = inquiryAccount.MetadataObj.SnapCoreResponse.BeneficiaryAccountName
	}

	return inquiryAccount, nil
}
