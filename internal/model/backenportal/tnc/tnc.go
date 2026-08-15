package tnc

import (
	"database/sql"
	"html/template"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
)

// TNC maps to the tncs table. Represents a managed TNC document.
type TNC struct {
	UUID            string       `json:"uuid" db:"uuid"`
	Version         string       `json:"version" db:"version"`
	Title           string       `json:"title" db:"title"`
	MarkdownContent string       `json:"markdownContent" db:"markdown_content"`
	IsActive        bool         `json:"isActive" db:"is_active"`
	CreatedBy       string       `json:"createdBy" db:"created_by"`
	CreatedAt       time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time    `json:"updatedAt" db:"updated_at"`
	DeletedAt       sql.NullTime `json:"deletedAt" db:"deleted_at"`
}

// MerchantTNCSigningHistory maps to the merchant_tnc_signing_histories table.
// Records each signing/acceptance event triggered by a user.
type MerchantTNCSigningHistory struct {
	UUID          string    `json:"uuid" db:"uuid"`
	MerchantID    string    `json:"merchantId" db:"merchant_id"`
	TNCVersionID  string    `json:"tncVersionId" db:"tnc_id"`
	Version       string    `json:"version" db:"version"`
	SignedBy      string    `json:"signedBy" db:"signed_by"`
	SignedByEmail string    `json:"signedByEmail" db:"signed_by_email"`
	SignedAt      time.Time `json:"signedAt" db:"signed_at"`
	DocumentPath  string    `json:"documentPath" db:"document_path"`
	IPAddress     string    `json:"ipAddress" db:"ip_address"`
	UserAgent     string    `json:"userAgent" db:"user_agent"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// CreateTNCVersionRequest is the CRM payload for creating a new TNC version.
type CreateTNCVersionRequest struct {
	Version       string `json:"version" validate:"required,max=50"`
	Title         string `json:"title" validate:"required,max=255"`
	HTMLContent   string `json:"htmlContent" validate:"required"`
	RequireResign bool   `json:"requireResign"`
	CreatedBy     string `json:"createdBy" validate:"required"`
}

// UpdateTNCVersionRequest is the CRM payload for updating an existing TNC version.
type UpdateTNCVersionRequest struct {
	ID            string  `json:"-" validate:"required"`
	Version       *string `json:"version" validate:"omitempty,max=50"`
	Title         *string `json:"title" validate:"omitempty,max=255"`
	HTMLContent   *string `json:"htmlContent"`
	RequireResign *bool   `json:"requireResign"`
}

// SignTNCRequest is the merchant-side payload for signing the active TNC.
type SignTNCRequest struct {
	MerchantID string `json:"-" validate:"required"`
	SignedBy   string `json:"-" validate:"required"`
	Email      string `json:"-" validate:"required"`
	IPAddress  string `json:"-"`
	UserAgent  string `json:"-"`
}

// TNCVersionQuery is the filter/pagination payload for listing TNC versions.
type TNCVersionQuery struct {
	Version  string `json:"version"`
	Title    string `json:"title"`
	IsActive *bool  `json:"isActive"`
	Page     int64  `json:"page"`
	PageSize int64  `json:"pageSize"`
	SortBy   string `json:"sortBy"`
	Sort     string `json:"sort"`
}

// SigningHistoryQuery is the filter/pagination payload for listing signing histories.
type SigningHistoryQuery struct {
	MerchantID string `json:"merchantId"`
	TNCVersion string `json:"version"`
	Page       int64  `json:"page"`
	PageSize   int64  `json:"pageSize"`
}

// TNCVersionResponse is the API response for a TNC version.
type TNCVersionResponse struct {
	ID            string    `json:"uuid"`
	Version       string    `json:"version"`
	Title         string    `json:"title"`
	HTMLContent   string    `json:"htmlContent"`
	RequireResign bool      `json:"requireResign"`
	IsActive      bool      `json:"isActive"`
	CreatedBy     string    `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// MerchantTNCSigningHistoryResponse is the API response for a signing event.
type MerchantTNCSigningHistoryResponse struct {
	ID            string    `json:"uuid"`
	MerchantID    string    `json:"merchantId"`
	TNCVersionID  string    `json:"tncVersionId"`
	Version       string    `json:"version"`
	SignedBy      string    `json:"signedBy"`
	SignedByEmail string    `json:"signedByEmail"`
	SignedAt      time.Time `json:"signedAt"`
	DocumentURL   string    `json:"documentUrl,omitempty"`
	IPAddress     string    `json:"ipAddress,omitempty"`
	UserAgent     string    `json:"userAgent,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// TNCSigningStatus describes whether a merchant needs to sign the active TNC.
type TNCSigningStatus struct {
	IsSigned        bool      `json:"isSigned"`
	ActiveVersion   string    `json:"activeVersion"`
	SignedVersion   string    `json:"signedVersion"`
	SignedAt        time.Time `json:"signedAt,omitzero"`
	SignedBy        string    `json:"signedBy,omitempty"`
	SignedByEmail   string    `json:"signedByEmail,omitempty"`
	MarkdownContent string    `json:"markdown_content,omitempty"`
	DocumentURL     string    `json:"documentUrl,omitempty"`
}

type TNCPDFTemplateData struct {
	PivotLogo     string
	MerchantName  string
	MerchantID    string
	Version       string
	Title         string
	TNCContent    template.HTML
	SignedByName  string
	SignedByEmail string
	SignedAt      string
	IPAddress     string
	UserAgent     string
}

func (r *CreateTNCVersionRequest) ToTNCVersion() *TNC {
	return &TNC{
		UUID:            util.GenerateUUID().String(),
		Version:         r.Version,
		Title:           r.Title,
		MarkdownContent: r.HTMLContent,
		IsActive:        false,
		CreatedBy:       r.CreatedBy,
	}
}

func (v *TNC) ApplyUpdate(req *UpdateTNCVersionRequest) {
	if req.Title != nil {
		v.Title = *req.Title
	}
	if req.HTMLContent != nil {
		v.MarkdownContent = *req.HTMLContent
	}
}

func (v *TNC) ToResponse() *TNCVersionResponse {
	return &TNCVersionResponse{
		ID:          v.UUID,
		Version:     v.Version,
		Title:       v.Title,
		HTMLContent: v.MarkdownContent,
		IsActive:    v.IsActive,
		CreatedBy:   v.CreatedBy,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func (h *MerchantTNCSigningHistory) ToResponse() *MerchantTNCSigningHistoryResponse {
	return &MerchantTNCSigningHistoryResponse{
		ID:            h.UUID,
		MerchantID:    h.MerchantID,
		TNCVersionID:  h.TNCVersionID,
		Version:       h.Version,
		SignedBy:      h.SignedBy,
		SignedByEmail: h.SignedByEmail,
		SignedAt:      h.SignedAt,
		IPAddress:     h.IPAddress,
		UserAgent:     h.UserAgent,
		CreatedAt:     h.CreatedAt,
	}
}

type SigningArtifact struct {
	DocumentPath string // GCS object path of the uploaded PDF (e.g. "tnc-documents/<mid>/<ver>/<uuid>.pdf")
}

// NewSigningHistory builds a signing history record from a sign request, the
// active TNC version, and the optional PDF artifact. When artifact is nil the
// history is recorded without a stored document — kept for compatibility with
// the prior MVP flow.
func NewSigningHistory(req *SignTNCRequest, version *TNC, artifact *SigningArtifact) *MerchantTNCSigningHistory {
	history := &MerchantTNCSigningHistory{
		UUID:          util.GenerateUUID().String(),
		MerchantID:    req.MerchantID,
		TNCVersionID:  version.UUID,
		Version:       version.Version,
		SignedBy:      req.SignedBy,
		SignedByEmail: req.Email,
		SignedAt:      time.Now().UTC(),
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
	}
	if artifact != nil {
		history.DocumentPath = artifact.DocumentPath
	}
	return history
}

func (q *TNCVersionQuery) BuildCondition() (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.Version != "" {
		conditions = append(conditions, "t.version = ?")
		args = append(args, q.Version)
	}
	if q.Title != "" {
		conditions = append(conditions, "t.title LIKE ?")
		args = append(args, "%"+q.Title+"%")
	}
	if q.IsActive != nil {
		conditions = append(conditions, "t.is_active = ?")
		args = append(args, *q.IsActive)
	}

	return strings.Join(conditions, " AND "), args
}

func (q *TNCVersionQuery) BuildOrderBy() string {
	allowedSortBy := map[string]string{
		"version":   "version",
		"title":     "title",
		"isActive":  "is_active",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}

	column, ok := allowedSortBy[q.SortBy]
	if !ok {
		return "created_at DESC"
	}

	sort := strings.ToUpper(q.Sort)
	if sort != "ASC" && sort != "DESC" {
		sort = "DESC"
	}

	return column + " " + sort
}

func (q *SigningHistoryQuery) BuildCondition() (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.MerchantID != "" {
		if _, err := uuid.Parse(q.MerchantID); err == nil {
			conditions = append(conditions, "t.merchant_id = ?")
			args = append(args, q.MerchantID)
		}
	}
	if q.TNCVersion != "" {
		conditions = append(conditions, "t.version = ?")
		args = append(args, q.TNCVersion)
	}

	return strings.Join(conditions, " AND "), args
}
