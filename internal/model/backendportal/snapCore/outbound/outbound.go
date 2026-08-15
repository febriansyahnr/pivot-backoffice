package outbound

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	OUTBOUND_TITLE_INQUIRY_STATUS                            = "INQUIRY_STATUS"
	OUTBOUND_TITLE_INQUIRY_STATUS_VA                         = "INQUIRY_STATUS_VA"
	OUTBOUND_TITLE_INQUIRY_STATUS_BIFAST                     = "INQUIRY_STATUS_BIFAST"
	OUTBOUND_TITLE_INQUIRY_ACCOUNT_INTERNAL                  = "INQUIRY_ACCOUNT_INTERNAL"
	OUTBOUND_TITLE_INQUIRY_ACCOUNT_EXTERNAL                  = "INQUIRY_ACCOUNT_EXTERNAL"
	OUTBOUND_TITLE_INQUIRY_ACCOUNT_EXTERNAL_BIFAST           = "INQUIRY_ACCOUNT_EXTERNAL_BIFAST"
	OUTBOUND_TITLE_INQUIRY_ACCOUNT_EXTERNAL_RTGS             = "INQUIRY_ACCOUNT_EXTERNAL_RTGS"
	OUTBOUND_TITLE_INQUIRY_ACCOUNT_EXTERNAL_SKN              = "INQUIRY_ACCOUNT_EXTERNAL_SKN"
	OUTBOUND_TITLE_CREATE_VA                                 = "CREATE_VA"
	OUTBOUND_TITLE_UPDATE_VA                                 = "UPDATE_VA"
	OUTBOUND_TITLE_INTRABANK_TRANSFER                        = "INTRABANK_TRANSFER"
	OUTBOUND_TITLE_INTRABANK_VA_TRANSFER                     = "INTRABANK_VA_TRANSFER"
	OUTBOUND_TITLE_INTERBANK_TRANSFER                        = "INTERBANK_TRANSFER"
	OUTBOUND_TITLE_INTERBANK_TRANSFER_BIFAST                 = "INTERBANK_TRANSFER_BIFAST"
	OUTBOUND_TITLE_RTGS_TRANSFER                             = "RTGS_TRANSFER"
	OUTBOUND_TITLE_SKN_TRANSFER                              = "SKN_TRANSFER"
	OUTBOUND_TITLE_B2B_REQUEST_TOKEN                         = "B2B_REQUEST_TOKEN"
	OUTBOUND_TITLE_B2B_REQUEST_SIGNATURE                     = "B2B_REQUEST_SIGNATURE"
	OUTBOUND_TITLE_BALANCE_INQUIRY                           = "BALANCE_INQUIRY"
	OUTBOUND_TITLE_QR_MPM_GENERATE                           = "QR_MPM_GENERATE"
	OUTBOUND_TITLE_QR_MPM_QUERY_PAYMENT                      = "QR_MPM_QUERY_PAYMENT"
	OUTBOUND_TITLE_QR_MPM_PAYMENT_STATUS                     = "QR_MPM_PAYMENT_STATUS"
	OUTBOUND_TITLE_QR_MPM_CANCELLATION                       = "QR_MPM_CANCELLATION"
	OUTBOUND_TITLE_QR_MPM_QUERY_PROVINCE                     = "QR_MPM_LIST_PROVINCE"
	OUTBOUND_TITLE_QR_MPM_QUERY_CITY                         = "QR_MPM_LIST_CITY"
	OUTBOUND_TITLE_QR_MPM_QUERY_DISTRCT                      = "QR_MPM_LIST_DISTRICT"
	OUTBOUND_TITLE_QR_ISSUING_INQUIRY                        = `QR_ISSUING_INQUIRY`
	OUTBOUND_TITLE_QR_ISSUING_REFUND                         = `QR_ISSUING_REFUND`
	OUTBOUND_TITLE_QR_ISSUING_PAYMENT                        = `QR_ISSUING_PAYMENT`
	OUTBOUND_TITLE_QR_ISSUING_STATUS_INQUIRY                 = `QR_ISSUING_STATUS_INQUIRY`
	OUTBOUND_TITLE_QR_MPM_QUERY_POSTCODE                     = "QR_MPM_LIST_POSTCODE"
	OUTBOUND_TITLE_QR_REGISTER_SUBMERCHANT_FRANCHISEE        = "QR_MPM_REGISTER_SUBMERCHANT_FRANCHISEE"
	OUTBOUND_TITLE_QR_REGISTER_DIRECT_STORE                  = "QR_MPM_REGISTER_DIRECT_STORE"
	OUTBOUND_TITLE_QR_REGISTER_INQUIRY                       = "QR_MPM_REGISTER_INQUIRY"
	OUTBOUND_TITLE_QR_IMAGE_UPLOAD                           = "QR_MPM_IMAGE_UPLOAD"
	OUTBOUND_TITLE_BANK_STATEMENT                            = "BANK_STATEMENT"
	OUTBOUND_TITLE_QR_MPM_UPDATE_MERCHANT_INFO               = "QR_MPM_UPDATE_MERCHANT_INFO"
	OUTBOUND_TITLE_QR_MPM_TRANSACTION_HISTORY                = "QR_MPM_TRANSACTION_HISTORY"
	OUTBOUND_TITLE_QR_MPM_REFUND                             = "QR_MPM_REFUND"
	OUTBOUND_TITLE_INQUIRY_WALLET_INCOMING_TRANSACTION_LIMIT = "INQUIRY_WALLET_INCOMING_TRANSACTION_LIMIT"
	OUTBOUND_TITLE_GET_BANK_LIST_RTGS                        = "GET_BANK_LIST_RTGS"
	OUTBOUND_TITLE_DIRECT_DEBIT                              = "DIRECT_DEBIT"
	OUTBOUND_TITLE_DEBIT_STATUS                              = "DEBIT_STATUS"
	OUTBOUND_TITLE_DEBIT_REFUND                              = "DEBIT_REFUND"
	OUTBOUND_TITLE_GET_ACCOUNT_TOPUP                         = "GET_ACCOUNT_TOPUP"
	OUTBOUND_TITLE_INQUIRY_VIRTUAL_ACCOUNT_INTERNAL          = "INQUIRY_VIRTUAL_ACCOUNT_INTERNAL"
	OUTBOUND_GET_AUTH_CODE                                   = "GET_AUTH_CODE"
	OUTBOUND_TITLE_ACCOUNT_BINDING                           = "ACCOUNT_BINDING"
)

type Outbound struct {
	UUID            string         `db:"uuid"`
	Title           string         `db:"title"`
	Acquirer        string         `db:"acquirer"`
	TransactionID   sql.NullString `json:"transaction_id" db:"transaction_id"`
	OriginID        sql.NullString `json:"origin_id" db:"origin_id"`
	OriginType      sql.NullString `json:"origin_type" db:"origin_type"`
	RequestPayload  sql.NullString `db:"request_payload"`
	ResponsePayload sql.NullString `db:"response_payload"`
	AdditionalInfo  sql.NullString `db:"additional_info"`
	CreatedAt       time.Time      `db:"created_at"`
}
type CreateOutboundParams struct {
	Title           string                  `json:"title" db:"title"`
	Acquirer        string                  `json:"acquirer" db:"acquirer"`
	TransactionID   sql.NullString          `json:"transaction_id" db:"transaction_id"`
	OriginID        sql.NullString          `json:"origin_id" db:"origin_id"`
	OriginType      sql.NullString          `json:"origin_type" db:"origin_type"`
	RequestPayload  interface{}             `json:"request_payload" db:"request_payload"`
	ResponsePayload interface{}             `json:"response_payload" db:"response_payload"`
	AdditionalInfo  *OutboundAdditionalInfo `json:"additional_info" db:"additional_info"`
}

type OutboundAdditionalInfo struct {
	TraceID string      `json:"trace_id" db:"trace_id,omitempty"`
	Header  interface{} `json:"header" db:"header,omitempty"`
	URL     string      `json:"url" db:"url,omitempty"`
	Info    *string     `json:"info" db:"info,omitempty"`
}

func CreateOutbound(params CreateOutboundParams) Outbound {
	outboundTitle := strings.ToLower(params.Title)

	responsePayload, _ := json.Marshal(params.ResponsePayload)

	res := Outbound{
		UUID:          uuid.NewString(),
		Title:         outboundTitle,
		Acquirer:      params.Acquirer,
		TransactionID: params.TransactionID,
		OriginID:      params.OriginID,
		OriginType:    params.OriginType,
		CreatedAt:     time.Now(),
	}

	if params.AdditionalInfo != nil {
		js, _ := json.Marshal(params.AdditionalInfo)

		res.AdditionalInfo = sql.NullString{
			String: string(js),
			Valid:  true,
		}
	}

	if params.RequestPayload != nil {
		requestPayload, _ := json.Marshal(params.RequestPayload)
		res.RequestPayload = sql.NullString{
			String: string(requestPayload),
			Valid:  true,
		}
	}

	if responsePayload != nil {
		res.ResponsePayload = sql.NullString{
			String: string(responsePayload),
			Valid:  true,
		}
	}

	return res
}

type OutboundQuery struct {
	CreatedAt string
	Title     string
	Acquirer  string
}

func (oq *OutboundQuery) String() string {
	queryArray := []string{}

	if oq.CreatedAt != "" {
		queryArray = append(queryArray, "date(created_at) = '"+oq.CreatedAt+"'")
	}

	if oq.Title != "" {
		queryArray = append(queryArray, "title = '"+oq.Title+"'")
	}

	if oq.Acquirer != "" {
		queryArray = append(queryArray, "acquirer = '"+oq.Acquirer+"'")
	}

	return strings.Join(queryArray, " AND ")
}
