# Vendor Management API - Requirements Document

## 1. Overview

### 1.1 Problem Statement

The Dashboard Vendor Management feature for **Card Funded Payout** requires backend APIs to manage vendor data used in the payment process.

Based on dashboard requirements, users must be able to:
- View vendor list
- Register new vendors
- View vendor details
- Update vendor data

Currently, the Vendor Management API is not available, preventing the dashboard from performing these operations.

### 1.2 Solution

Create a **Vendor Management API** that provides capabilities to:
- Create new vendor data
- Retrieve vendor list
- View vendor details
- Update vendor data
- Delete or deactivate vendors (soft delete)

This API will be used by the **BE Portal Dashboard** to support vendor management processes in the Card Funded Payout feature.

---

## 2. Functional Requirements

### 2.1 Create Vendor

| Attribute | Description |
|-----------|-------------|
| **Endpoint** | `POST /crm/v1/card-funded-payout/vendors` |
| **Description** | Register a new vendor in the system |
| **Access** | CRM Dashboard (Internal) |
| **Auth** | X-CRM-Key header |

**Request Payload:**
```json
{
  "name": "string",
  "beneficialOwner": "string",
  "businessCategory": "string",
  "avgMonthlyTpvAmount": 1000000,
  "bankName": "string",
  "bankCode": "string",
  "accountNumber": "string",
  "accountName": "string",
  "documents": [
    {
      "type": "KTP",
      "external": "card-funded-payout/vendors/img/abc123.pdf",
      "internal": {
        "bucket": "BUCKET_NAME",
        "object": "card-funded-payout/vendors/uuid/KTP-1722494560.pdf"
      }
    }
  ]
}
```

> **Note:** `documents` field is **optional**. Documents should be uploaded to GCS first before creating/updating vendor.

**Response:**
```json
{
  "code": "OK",
  "message": "Vendor created successfully",
  "data": {
    "id": "uuid",
    "name": "string",
    "status": "ACTIVE",
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### 2.2 Get Vendor List

| Attribute | Description |
|-----------|-------------|
| **Endpoint** | `GET /crm/v1/card-funded-payout/vendors` |
| **Description** | Retrieve paginated list of vendors |
| **Access** | CRM Dashboard (Internal) |
| **Auth** | X-CRM-Key header |

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | integer | No | Page number (default: 1) |
| `perPage` | integer | No | Items per page (default: 10, max: 100) |
| `status` | string | No | Filter by status (ACTIVE, INACTIVE) |
| `name` | string | No | Search by vendor name |

**Response:**
```json
{
  "code": "OK",
  "message": "Success",
  "data": [
    {
      "id": "uuid",
      "name": "string",
      "beneficialOwner": "string",
      "businessCategory": "string",
      "avgMonthlyTpvAmount": 1000000,
      "bankName": "string",
      "bankCode": "string",
      "accountNumber": "string",
      "accountName": "string",
      "status": "ACTIVE",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "perPage": 10,
    "totalData": 100,
    "totalPage": 10
  }
}
```

---

### 2.3 Get Vendor Detail

| Attribute | Description |
|-----------|-------------|
| **Endpoint** | `GET /crm/v1/card-funded-payout/vendors/{id}` |
| **Description** | Retrieve detailed information of a specific vendor |
| **Access** | CRM Dashboard (Internal) |
| **Auth** | X-CRM-Key header |

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Vendor ID |

**Response:**
```json
{
  "code": "OK",
  "message": "Success",
  "data": {
    "id": "uuid",
    "name": "string",
    "beneficialOwner": "string",
    "businessCategory": "string",
    "avgMonthlyTpvAmount": 1000000,
    "bankName": "string",
    "bankCode": "string",
    "accountNumber": "string",
    "accountName": "string",
    "documents": [
      {
        "type": "KTP",
        "external": "card-funded-payout/vendors/img/abc123.pdf",
        "internal": {
          "bucket": "BUCKET_NAME",
          "object": "card-funded-payout/vendors/uuid/KTP-1722494560.pdf"
        }
      }
    ],
    "status": "ACTIVE",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### 2.4 Update Vendor

| Attribute | Description |
|-----------|-------------|
| **Endpoint** | `PUT /crm/v1/card-funded-payout/vendors/{id}` |
| **Description** | Update existing vendor data |
| **Access** | CRM Dashboard (Internal) |
| **Auth** | X-CRM-Key header |

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Vendor ID |

**Request Payload:**
```json
{
  "name": "string",
  "beneficialOwner": "string",
  "businessCategory": "string",
  "avgMonthlyTpvAmount": 1000000,
  "bankName": "string",
  "bankCode": "string",
  "accountNumber": "string",
  "accountName": "string",
  "documents": [
    {
      "type": "KTP",
      "external": "card-funded-payout/vendors/img/abc123.pdf",
      "internal": {
        "bucket": "BUCKET_NAME",
        "object": "card-funded-payout/vendors/uuid/KTP-1722494560.pdf"
      }
    }
  ]
}
```

> **Note:** `documents` field is **optional**.

**Response:**
```json
{
  "code": "OK",
  "message": "Vendor updated successfully",
  "data": {
    "id": "uuid",
    "name": "string",
    "status": "ACTIVE",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

---

### 2.5 Delete Vendor (Soft Delete)

| Attribute | Description |
|-----------|-------------|
| **Endpoint** | `DELETE /crm/v1/card-funded-payout/vendors/{id}` |
| **Description** | Soft delete a vendor (set deleted_at timestamp) |
| **Access** | CRM Dashboard (Internal) |
| **Auth** | X-CRM-Key header |

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Vendor ID |

**Response:**
```json
{
  "code": "OK",
  "message": "Vendor deleted successfully",
  "data": null
}
```

---

### 2.6 Upload Vendor Document (Optional - Future Enhancement)

| Attribute | Description |
|-----------|-------------|
| **Endpoint** | `POST /crm/v1/card-funded-payout/vendors/{id}/documents` |
| **Description** | Upload document for a vendor to GCS |
| **Access** | CRM Dashboard (Internal) |
| **Auth** | X-CRM-Key header |
| **Content-Type** | multipart/form-data |

**Request (multipart/form-data):**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Document type (KTP, NPWP, SIUP, etc.) |
| `file` | file | Yes | Document file |

**Response:**
```json
{
  "code": "OK",
  "message": "Document uploaded successfully",
  "data": {
    "type": "KTP",
    "external": "card-funded-payout/vendors/img/abc123.pdf",
    "internal": {
      "bucket": "BUCKET_NAME",
      "object": "card-funded-payout/vendors/uuid/KTP-1722494560.pdf"
    }
  }
}
```

> **Note:** This endpoint follows the merchant document upload pattern (`internal/service/v1/merchant/document.go`).

---

## 3. Data Model

### 3.1 Database Schema

**Table: `vendors`**

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| `uuid` | VARCHAR(36) | NO | Primary key (UUID v7) |
| `name` | VARCHAR(255) | NO | Vendor name |
| `beneficial_owner` | VARCHAR(255) | NO | Beneficial owner name |
| `business_category` | VARCHAR(100) | NO | Business category |
| `avg_monthly_tpv_amount` | DECIMAL(20,2) | NO | Average monthly TPV amount |
| `bank_name` | VARCHAR(100) | NO | Bank name |
| `bank_code` | VARCHAR(20) | NO | Bank code |
| `account_number` | VARCHAR(50) | NO | Bank account number |
| `account_name` | VARCHAR(255) | NO | Account holder name |
| `documents` | JSON | YES | Supporting documents |
| `status` | ENUM('ACTIVE', 'INACTIVE') | NO | Vendor status |
| `created_at` | TIMESTAMP | NO | Creation timestamp |
| `updated_at` | TIMESTAMP | NO | Last update timestamp |
| `deleted_at` | TIMESTAMP | YES | Soft delete timestamp |

**Indexes:**
- PRIMARY KEY (`uuid`)
- INDEX `idx_status` (`status`)
- INDEX `idx_name` (`name`)
- INDEX `idx_created_at` (`created_at`)
- INDEX `idx_deleted_at` (`deleted_at`)

### 3.2 Documents JSON Structure

Documents are **optional**. When provided, they should be uploaded to GCS first, then stored with the following structure:

```json
[
  {
    "type": "KTP",
    "external": "card-funded-payout/vendors/img/esaq_7CYytg0kBgDQrgmlDn5k5oTPdPwc0AJzGASkTs.pdf",
    "internal": {
      "bucket": "BUCKET_NAME",
      "object": "card-funded-payout/vendors/93cc50b8-f8e6-4771-a808-bf397c0be568/KTP-1722494560.pdf"
    }
  },
  {
    "type": "NPWP",
    "external": "card-funded-payout/vendors/img/xyz_8DZzuH1kBgDQrgmlDn5k5oTPdPwc0AJzGASkTs.pdf",
    "internal": {
      "bucket": "BUCKET_NAME",
      "object": "card-funded-payout/vendors/93cc50b8-f8e6-4771-a808-bf397c0be568/NPWP-1722494560.pdf"
    }
  }
]
```

**Document Upload Flow** (following merchant document pattern):
1. Client uploads file via multipart/form-data to a separate upload endpoint
2. Service uploads file to GCS using `gcs.UploadFileFromMultipart()`
3. GCS returns bucket and object name
4. Store document metadata with internal (bucket/object) and external (signed URL path) references
5. When retrieving, generate signed URL using `gcs.CreateSignedURL()`

---

## 4. Test Scenarios (Given-When-Then)

### Scenario 1: Create Vendor

| Step | Description |
|------|-------------|
| **Given** | User is on the Register Vendor page |
| **When** | User fills in vendor data and submits |
| **Then** | New vendor is successfully created and stored in the system |

**Acceptance Criteria:**
- [ ] Vendor name is required and must not be empty
- [ ] Beneficial owner is required
- [ ] Business category is required
- [ ] Bank account details (bank_name, bank_code, account_number, account_name) are required
- [ ] avg_monthly_tpv_amount must be a positive number
- [ ] System generates UUID v7 for vendor
- [ ] System records created_at and updated_at timestamp
- [ ] Status defaults to ACTIVE

---

### Scenario 2: View Vendor List

| Step | Description |
|------|-------------|
| **Given** | Vendor data exists in the system |
| **When** | User opens the Vendor List page |
| **Then** | System displays available vendor list |

**Acceptance Criteria:**
- [ ] List is paginated (default 10 items per page)
- [ ] List can be filtered by status
- [ ] List can be searched by vendor name
- [ ] List shows only non-deleted vendors (deleted_at IS NULL)

---

### Scenario 3: View Vendor Detail

| Step | Description |
|------|-------------|
| **Given** | Vendor is registered |
| **When** | User selects a vendor |
| **Then** | System displays vendor detail information |

**Acceptance Criteria:**
- [ ] All vendor information is displayed
- [ ] Bank account details are displayed
- [ ] Documents are listed
- [ ] Returns 404 if vendor not found

---

### Scenario 4: Update Vendor

| Step | Description |
|------|-------------|
| **Given** | Vendor is registered |
| **When** | User makes changes to vendor data |
| **Then** | Vendor data changes are successfully saved |

**Acceptance Criteria:**
- [ ] All editable fields can be updated
- [ ] System records updated_at timestamp
- [ ] Returns 404 if vendor not found
- [ ] Cannot update deleted vendor

---

### Scenario 5: Delete Vendor (Soft Delete)

| Step | Description |
|------|-------------|
| **Given** | Vendor is registered |
| **When** | User deletes the vendor |
| **Then** | Vendor is soft deleted (deleted_at is set) |

**Acceptance Criteria:**
- [ ] Soft delete (set deleted_at and updated_at timestamp)
- [ ] Returns 404 if vendor not found
- [ ] Deleted vendor does not appear in list

---

## 5. Error Handling

| Error Code | HTTP Status | Message | Description |
|------------|-------------|---------|-------------|
| `vendor_not_found` | 404 | Vendor not found | Vendor ID does not exist |
| `invalid_request_payload` | 400 | Invalid request payload | Malformed JSON |
| `validation_error` | 400 | Validation failed | Field validation error |
| `invalid_id` | 400 | Invalid ID format | UUID format invalid |
| `internal_error` | 500 | Internal server error | Unexpected server error |

---

## 6. Architecture

### 6.1 Layer Structure (Following fraud-rule Pattern)

```
port/http/controller/v1/crmController/cardFundedPayoutVendor/
├── type.go           # Controller struct and constructor
├── create.go         # POST /crm/v1/card-funded-payout/vendors
├── list.go           # GET /crm/v1/card-funded-payout/vendors
├── detail.go         # GET /crm/v1/card-funded-payout/vendors/{id}
├── update.go         # PUT /crm/v1/card-funded-payout/vendors/{id}
├── delete.go         # DELETE /crm/v1/card-funded-payout/vendors/{id}
├── create_test.go
├── list_test.go
├── detail_test.go
├── update_test.go
└── delete_test.go

internal/service/v1/cardFundedPayoutVendor/
├── type.go           # Service struct and constructor
├── create.go
├── list.go
├── detail.go
├── update.go
├── delete.go
└── *_test.go

internal/repository/cardFundedPayoutVendor/
├── type.go           # Repository struct and constructor
├── create.go         # INSERT query
├── get.go            # SELECT queries (list + detail)
├── update.go         # UPDATE query
├── delete.go         # Soft DELETE (UPDATE deleted_at)
└── *_test.go

internal/model/cardFundedPayoutVendor/
├── vendor.go         # Entity, Request, Response, Query structs
```

### 6.2 Route Registration

Add to `port/http/route.go` under `/crm/v1` route (after line ~1578):

```go
// card-funded-payout vendors (following fraud-rule pattern)
r.Route("/card-funded-payout", func(r chi.Router) {
    r.Route("/vendors", func(r chi.Router) {
        r.Post("/", module.V1CRMCardFundedPayoutVendorController.Create)
        r.Put("/{id}", module.V1CRMCardFundedPayoutVendorController.Update)
        r.Delete("/{id}", module.V1CRMCardFundedPayoutVendorController.Delete)
        r.Get("/{id}", module.V1CRMCardFundedPayoutVendorController.Detail)
        r.Get("/", module.V1CRMCardFundedPayoutVendorController.List)
    })
})
```

### 6.3 Interface Definitions

**Repository Interface** (add to `internal/repository/repository.go`):
```go
type ICardFundedPayoutVendorRepository interface {
    Create(ctx context.Context, vendor *cardFundedPayoutVendorModel.Vendor) error
    Update(ctx context.Context, vendor *cardFundedPayoutVendorModel.Vendor) error
    Delete(ctx context.Context, uuid string) error
    GetByID(ctx context.Context, uuid string) (*cardFundedPayoutVendorModel.Vendor, error)
    List(ctx context.Context, q *cardFundedPayoutVendorModel.VendorQuery) ([]*cardFundedPayoutVendorModel.Vendor, int, error)
}
```

**Service Interface** (add to `internal/service/service.go`):
```go
type ICardFundedPayoutVendorService interface {
    Create(ctx context.Context, request *cardFundedPayoutVendorModel.CreateVendorRequest) (*cardFundedPayoutVendorModel.VendorResponse, error)
    Update(ctx context.Context, request *cardFundedPayoutVendorModel.UpdateVendorRequest) (*cardFundedPayoutVendorModel.VendorResponse, error)
    Delete(ctx context.Context, uuid string) error
    Detail(ctx context.Context, uuid string) (*cardFundedPayoutVendorModel.Vendor, error)
    List(ctx context.Context, req *cardFundedPayoutVendorModel.VendorQuery) (*commonModel.PaginationResponse, error)
}
```

**Controller Interface** (add to `port/http/controller/type.go`):
```go
type V1CRMCardFundedPayoutVendorController interface {
    Create(w http.ResponseWriter, r *http.Request)
    Update(w http.ResponseWriter, r *http.Request)
    Delete(w http.ResponseWriter, r *http.Request)
    Detail(w http.ResponseWriter, r *http.Request)
    List(w http.ResponseWriter, r *http.Request)
}
```

---

## 7. Model Definition

### 7.1 Entity (following fraud-rules pattern)

```go
package cardfundedpayoutvendormodel

import (
    "database/sql"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

type Vendor struct {
    UUID                 string          `json:"uuid" db:"uuid"`
    Name                 string          `json:"name" db:"name"`
    BeneficialOwner      string          `json:"beneficialOwner" db:"beneficial_owner"`
    BusinessCategory     string          `json:"businessCategory" db:"business_category"`
    AvgMonthlyTpvAmount  decimal.Decimal `json:"avgMonthlyTpvAmount" db:"avg_monthly_tpv_amount"`
    BankName             string          `json:"bankName" db:"bank_name"`
    BankCode             string          `json:"bankCode" db:"bank_code"`
    AccountNumber        string          `json:"accountNumber" db:"account_number"`
    AccountName          string          `json:"accountName" db:"account_name"`
    Documents            JSONDocuments   `json:"documents" db:"documents"`
    Status               string          `json:"status" db:"status"`
    CreatedAt            time.Time       `json:"createdAt" db:"created_at"`
    UpdatedAt            time.Time       `json:"updatedAt" db:"updated_at"`
    DeletedAt            sql.NullTime    `json:"deletedAt" db:"deleted_at"`
}

type JSONDocuments []Document

type Document struct {
    Type     string       `json:"type"`
    External string       `json:"external"`
    Internal *DocLocation `json:"internal"`
}

type DocLocation struct {
    Bucket string `json:"bucket"`
    Object string `json:"object"`
}

type VendorQuery struct {
    Name     string `json:"name"`
    Status   string `json:"status"`
    Page     int64  `json:"page"`
    PageSize int64  `json:"pageSize"`
}

type CreateVendorRequest struct {
    Name                string          `json:"name" validate:"required"`
    BeneficialOwner     string          `json:"beneficialOwner" validate:"required"`
    BusinessCategory    string          `json:"businessCategory" validate:"required"`
    AvgMonthlyTpvAmount decimal.Decimal `json:"avgMonthlyTpvAmount" validate:"required"`
    BankName            string          `json:"bankName" validate:"required"`
    BankCode            string          `json:"bankCode" validate:"required"`
    AccountNumber       string          `json:"accountNumber" validate:"required"`
    AccountName         string          `json:"accountName" validate:"required"`
    Documents           []Document      `json:"documents,omitempty"` // Optional - upload to GCS first
}

type UpdateVendorRequest struct {
    UUID                string           `json:"uuid" validate:"required"`
    Name                *string          `json:"name"`
    BeneficialOwner     *string          `json:"beneficialOwner"`
    BusinessCategory    *string          `json:"businessCategory"`
    AvgMonthlyTpvAmount *decimal.Decimal `json:"avgMonthlyTpvAmount"`
    BankName            *string          `json:"bankName"`
    BankCode            *string          `json:"bankCode"`
    AccountNumber       *string          `json:"accountNumber"`
    AccountName         *string          `json:"accountName"`
    Documents           *[]Document      `json:"documents"`
}

type VendorResponse struct {
    UUID                string          `json:"uuid"`
    Name                string          `json:"name"`
    BeneficialOwner     string          `json:"beneficialOwner"`
    BusinessCategory    string          `json:"businessCategory"`
    AvgMonthlyTpvAmount decimal.Decimal `json:"avgMonthlyTpvAmount"`
    BankName            string          `json:"bankName"`
    BankCode            string          `json:"bankCode"`
    AccountNumber       string          `json:"accountNumber"`
    AccountName         string          `json:"accountName"`
    Documents           []Document      `json:"documents,omitempty"` // Optional
    Status              string          `json:"status"`
    CreatedAt           time.Time       `json:"createdAt"`
    UpdatedAt           time.Time       `json:"updatedAt"`
    DeletedAt           *time.Time      `json:"deletedAt,omitempty"`
}

// VendorDocumentResponse for Get Detail with signed URLs
type VendorDocumentResponse struct {
    Type string `json:"type"`
    URL  string `json:"url"` // Signed URL from GCS
}

// Factory function
func New(req *CreateVendorRequest) (*Vendor, error) {
    id, err := uuid.NewV7()
    if err != nil {
        return nil, err
    }
    now := time.Now().UTC()
    return &Vendor{
        UUID:                id.String(),
        Name:                req.Name,
        BeneficialOwner:     req.BeneficialOwner,
        BusinessCategory:    req.BusinessCategory,
        AvgMonthlyTpvAmount: req.AvgMonthlyTpvAmount,
        BankName:            req.BankName,
        BankCode:            req.BankCode,
        AccountNumber:       req.AccountNumber,
        AccountName:         req.AccountName,
        Documents:           req.Documents,
        Status:              "ACTIVE",
        CreatedAt:           now,
        UpdatedAt:           now,
    }, nil
}

// Update method
func (v *Vendor) Update(req *UpdateVendorRequest) {
    if req.Name != nil {
        v.Name = *req.Name
    }
    if req.BeneficialOwner != nil {
        v.BeneficialOwner = *req.BeneficialOwner
    }
    if req.BusinessCategory != nil {
        v.BusinessCategory = *req.BusinessCategory
    }
    if req.AvgMonthlyTpvAmount != nil {
        v.AvgMonthlyTpvAmount = *req.AvgMonthlyTpvAmount
    }
    if req.BankName != nil {
        v.BankName = *req.BankName
    }
    if req.BankCode != nil {
        v.BankCode = *req.BankCode
    }
    if req.AccountNumber != nil {
        v.AccountNumber = *req.AccountNumber
    }
    if req.AccountName != nil {
        v.AccountName = *req.AccountName
    }
    if req.Documents != nil {
        v.Documents = *req.Documents
    }
    v.UpdatedAt = time.Now().UTC()
}

// ToResponse converter
func (v *Vendor) ToResponse() *VendorResponse {
    var deletedAt *time.Time
    if v.DeletedAt.Valid {
        deletedAt = &v.DeletedAt.Time
    }
    return &VendorResponse{
        UUID:                v.UUID,
        Name:                v.Name,
        BeneficialOwner:     v.BeneficialOwner,
        BusinessCategory:    v.BusinessCategory,
        AvgMonthlyTpvAmount: v.AvgMonthlyTpvAmount,
        BankName:            v.BankName,
        BankCode:            v.BankCode,
        AccountNumber:       v.AccountNumber,
        AccountName:         v.AccountName,
        Documents:           v.Documents,
        Status:              v.Status,
        CreatedAt:           v.CreatedAt,
        UpdatedAt:           v.UpdatedAt,
        DeletedAt:           deletedAt,
    }
}
```

---

## 8. Migration

```sql
-- Migration: Create vendors table
-- Version: YYYYMMDDHHMMSS_create_vendors_table

-- +goose Up
CREATE TABLE vendors (
    uuid VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    beneficial_owner VARCHAR(255) NOT NULL,
    business_category VARCHAR(100) NOT NULL,
    avg_monthly_tpv_amount DECIMAL(20,2) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(20) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    documents JSON NULL,
    status ENUM('ACTIVE', 'INACTIVE') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    PRIMARY KEY (uuid),
    INDEX idx_status (status),
    INDEX idx_name (name),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS vendors;
```

---

## 9. Implementation Checklist

| # | Task | Status |
|---|------|--------|
| 1 | Create database migration | [ ] |
| 2 | Create vendor model (internal/model/cardFundedPayoutVendor/vendor.go) | [ ] |
| 3 | Add repository interface to repository.go | [ ] |
| 4 | Create repository implementation | [ ] |
| 5 | Add service interface to service.go | [ ] |
| 6 | Create service implementation | [ ] |
| 7 | Add controller interface to controller/type.go | [ ] |
| 8 | Create controller handlers | [ ] |
| 9 | Register routes in route.go | [ ] |
| 10 | Wire dependencies in module.go | [ ] |
| 11 | Generate mocks (make gen-mocks) | [ ] |
| 12 | Write unit tests for repository | [ ] |
| 13 | Write unit tests for service | [ ] |
| 14 | Write unit tests for controller | [ ] |
| 15 | Add Swagger documentation | [ ] |

---

## 10. References

- **Fraud Rule Pattern**: `port/http/controller/v1/crmController/fraudRule/`
- **Fraud Rule Model**: `internal/model/fraudRules/fraudRules.go`
- **Fraud Rule Repository**: `internal/repository/fraudRules/`
- **Fraud Rule Service**: `internal/service/v1/fraudRule/`
- **CRM Routes**: `port/http/route.go` (line 1571-1578)
- **Repository Interfaces**: `internal/repository/repository.go`
- **Service Interfaces**: `internal/service/service.go`
- **Document Upload Pattern**: `internal/service/v1/merchant/document.go`
- **Document Model**: `internal/model/merchant/document.go` (DocLocation struct)
- **GCS Upload**: `gcs.UploadFileFromMultipart()`, `gcs.CreateSignedURL()`
