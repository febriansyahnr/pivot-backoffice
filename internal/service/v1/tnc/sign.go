package tnc

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/yuin/goldmark"
)

// SignTNC records the authenticated user's acceptance of the currently active
// TNC version on behalf of their merchant.
func (s *TNCService) SignTNC(
	ctx context.Context,
	req *tncModel.SignTNCRequest,
) (*tncModel.MerchantTNCSigningHistoryResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/tnc/SignTNC")
	defer span.End()

	activeTNC, err := s.repo.GetActiveTNCVersion(ctx)
	if err != nil {
		s.logger.Error(ctx, "error when getting active tnc version", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if activeTNC == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrNoActiveTNCVersion)
	}

	signedVersion, err := s.repo.GetSigningByMerchantAndVersion(ctx, req.MerchantID, activeTNC.Version)
	if err != nil {
		s.logger.Error(ctx, "error when checking existing tnc signing", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if signedVersion != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantAlreadySignedTNC)
	}

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, req.MerchantID)
	if err != nil || merchant == nil {
		s.logger.Error(ctx, "error when finding merchant for tnc signing", logger.Error(err), logger.String("merchantID", req.MerchantID))
		if err == nil {
			err = constant.ErrInvalidMerchantID
		}
		return nil, pkgErrs.New(response.HttpErrNotFound, err)
	}

	if merchant.ParentID.Valid && merchant.ParentID.String != "" {
		s.logger.Error(ctx, "merchant is a submerchant, cannot sign tnc", logger.String("merchantID", req.MerchantID))
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrSubmerchantCannotSignTNC)
	}

	artifact, err := s.producePDF(ctx, req, activeTNC, merchant)
	if err != nil {
		s.logger.Error(ctx, "error when producing tnc pdf", logger.Error(err), logger.String("merchantID", req.MerchantID))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}

	history := tncModel.NewSigningHistory(req, activeTNC, artifact)
	if err := s.repo.InsertSigningHistory(ctx, history); err != nil {
		s.logger.Error(ctx, "error when inserting tnc signing history", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	signedAt := history.SignedAt
	metaStatus := &merchantModel.TNCSigningMetadata{
		IsSigned:      true,
		SignedVersion: history.Version,
		SignedPath:    history.DocumentPath,
		SignedAt:      &signedAt,
		SignedBy:      history.SignedByEmail,
	}
	if err := merchant.UpdateTNCSigningStatus(metaStatus); err != nil {
		s.logger.Error(ctx, "error when updating merchant tnc metadata", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}
	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		s.logger.Error(ctx, "error when persisting merchant tnc metadata", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if s.activity != nil {
		s.recordActivity(ctx, req, activeTNC.Version)
	}

	resp := history.ToResponse()
	resp.DocumentURL = s.signedDocumentURL(ctx, history.DocumentPath)
	return resp, nil
}

func (s *TNCService) producePDF(
	ctx context.Context,
	req *tncModel.SignTNCRequest,
	activeTNC *tncModel.TNC,
	merchant *merchantModel.Merchant,
) (*tncModel.SigningArtifact, error) {
	if s.gcs == nil {
		return nil, nil
	}

	signedAt := time.Now().UTC()
	tncHTML, err := markdownToHTML(activeTNC.MarkdownContent)
	if err != nil {
		s.logger.Error(ctx, "failed to render tnc markdown to html", logger.Error(err), logger.String("merchantID", merchant.UUID), logger.String("tncVersion", activeTNC.Version))
		return nil, err
	}
	renderData := tncModel.TNCPDFTemplateData{
		PivotLogo:     s.cfg.MerchantPortalConfig.LogoURL,
		MerchantName:  merchant.Name,
		MerchantID:    merchant.UUID,
		Version:       activeTNC.Version,
		Title:         activeTNC.Title,
		TNCContent:    tncHTML,
		SignedByName:  req.SignedBy,
		SignedByEmail: req.Email,
		SignedAt:      signedAt.Format(time.RFC1123),
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
	}

	buf := new(bytes.Buffer)

	r := pdf.NewRequestPdf(wkhtmltopdf.PageSizeA4, wkhtmltopdf.OrientationPortrait, pdf.WithOutput(buf))
	err = r.GeneratePDF(ctx, TNCTemplatePath, renderData)
	if err != nil {
		s.logger.Error(ctx, "failed to generate tnc document",
			logger.Error(err),
			logger.String("merchantID", merchant.UUID),
			logger.String("tncVersion", activeTNC.Version),
			logger.String("signerEmail", req.SignedBy))
		return nil, err
	}

	objectName := path.Join(
		s.cfg.GCSConfig.TNCDocumentFolderName,
		req.MerchantID,
		activeTNC.Version,
		uuid.NewString()+".pdf",
	)
	reader := bytes.NewReader(buf.Bytes())
	if _, err := s.gcs.UploadFile(ctx, objectName, reader, true); err != nil {
		return nil, fmt.Errorf("upload tnc pdf to gcs: %w", err)
	}

	// Store the GCS object path; the signed URL is generated on read so the
	// URL never goes stale.
	return &tncModel.SigningArtifact{
		DocumentPath: objectName,
	}, nil
}

// recordActivity fires the activity log row. Errors are logged but not
// returned — by the time we get here the signing has already been committed.
func (s *TNCService) recordActivity(ctx context.Context, req *tncModel.SignTNCRequest, version string) {
	if s.activity == nil {
		return
	}
	now := time.Now().UTC()
	parameter := map[string]any{
		"version":    version,
		"merchantId": req.MerchantID,
		"ipAddress":  req.IPAddress,
	}
	userID := req.SignedBy
	if err := s.activity.Create(ctx, &activityModel.Activity{
		ID:          uuid.NewString(),
		MerchantID:  req.MerchantID,
		UserID:      &userID,
		Tag:         constant.TagTNC,
		Activity:    constant.ActivityMerchantTNCSign,
		ServiceName: "TNC Service",
		Parameter:   &parameter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		s.logger.Error(ctx, "non-critical: failed to record tnc signing activity", logger.Error(err))
	}
}

// markdownToHTML renders the stored TNC markdown into HTML safe for embedding
// in the PDF template. Goldmark is CommonMark compliant; the output is already
// HTML-escaped so it is safe to wrap in template.HTML (skips Go's auto-escaping
// so the markup is rendered, not the raw tags).
func markdownToHTML(md string) (template.HTML, error) {
	if md == "" {
		return "", nil
	}
	var buf bytes.Buffer
	if err := goldmark.New().Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("render tnc markdown: %w", err)
	}
	//nolint:gosec // G203: md is admin-authored TNC content from tncs.markdown_content,
	// not end-user input. Goldmark HTML-escapes inline text/code, so the output is safe
	// to render as HTML in the PDF template (wrapping in template.HTML is the correct way
	// to embed pre-rendered HTML in html/template so the tags render instead of escaping).
	return template.HTML(buf.String()), nil
}
