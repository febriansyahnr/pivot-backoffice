package tnc

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

// GetTNCSigningStatus reports whether a merchant has signed the currently
// active TNC version. If the active version requires re-signing and the
// merchant's recorded version differs from the active one, isSigned=false.
func (s *TNCService) GetTNCSigningStatus(ctx context.Context, merchantID string) (*tncModel.TNCSigningStatus, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/GetTNCSigningStatus")
	defer span.End()

	status := &tncModel.TNCSigningStatus{
		IsSigned:      false,
		ActiveVersion: "",
		SignedVersion: "",
	}

	active, err := s.repo.GetActiveTNCVersion(ctx)
	if err != nil {
		s.logger.Error(ctx, "error when getting active tnc version", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if active == nil {
		s.logger.Warn(ctx, "active tnc version not found")
		return nil, nil
	}

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "error when finding merchant by id", logger.Error(err), logger.String("merchantID", merchantID))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if merchant == nil {
		s.logger.Error(ctx, "merchant not found", logger.String("merchantID", merchantID))
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	if merchant.ParentID.Valid && merchant.ParentID.String != "" {
		s.logger.Warn(ctx, "merchant is a submerchant, cannot get tnc signing status", logger.String("merchantID", merchantID))
		return nil, nil
	}

	latest, err := s.repo.GetLatestSigningByMerchant(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "error when getting latest tnc signing for merchant", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if latest != nil {
		status.ActiveVersion = active.Version
		status.SignedVersion = latest.Version
		status.SignedAt = latest.SignedAt
		status.SignedBy = latest.SignedBy
		status.SignedByEmail = latest.SignedByEmail
		status.DocumentURL = s.signedDocumentURL(ctx, latest.DocumentPath)
		status.IsSigned = latest.Version == active.Version
	}

	if !status.IsSigned {
		status.MarkdownContent = active.MarkdownContent
	}

	return status, nil
}

func (s *TNCService) GetSigningHistory(
	ctx context.Context,
	q *tncModel.SigningHistoryQuery,
) (*commonModel.PaginationResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/GetSigningHistory")
	defer span.End()

	if q.MerchantID == "" {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidMerchantID)
	}

	list, total, err := s.repo.ListSigningHistories(ctx, q)
	if err != nil {
		s.logger.Error(ctx, "error when listing tnc signing histories", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrGetTNCSigningHistory)
	}

	responses := make([]*tncModel.MerchantTNCSigningHistoryResponse, 0, len(list))
	for _, h := range list {
		resp := h.ToResponse()
		resp.DocumentURL = s.signedDocumentURL(ctx, h.DocumentPath)
		responses = append(responses, resp)
	}

	return &commonModel.PaginationResponse{
		Data: responses,
		Meta: *commonModel.NewMeta(q.Page, q.PageSize, int64(total)),
	}, nil
}

// signedDocumentURL generates a fresh GCS signed URL for the given object
// path. Returns an empty string when path is empty or the URL cannot be
// generated, so callers always get a usable-or-empty value rather than an
// error that would block the whole history response.
func (s *TNCService) signedDocumentURL(ctx context.Context, documentPath string) string {
	if documentPath == "" || s.gcs == nil {
		return ""
	}
	url, err := s.gcs.CreateSignedURL(ctx, documentPath, TNCSignedURLTTL)
	if err != nil {
		s.logger.Error(ctx, "failed to generate tnc document signed url", logger.Error(err), logger.String("documentPath", documentPath))
		return ""
	}
	return url
}
