package disbursementModel

import (
	"encoding/json"
	"time"

	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	snapCoreModelBT "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type BulkPreviewResponse struct {
	ReferenceID            string   `json:"referenceId"`
	BeneficiaryBankCode    string   `json:"beneficiaryBankCode"`
	BeneficiaryBankName    string   `json:"beneficiaryBankName"`
	BeneficiaryAccountNo   string   `json:"beneficiaryAccountNo"`
	BeneficiaryAccountName string   `json:"beneficiaryAccountName"`
	FeeAmount              *float64 `json:"feeAmount,omitempty"`
	Amount                 string   `json:"amount"`
	Remark                 string   `json:"remark"`
	Error                  string   `json:"error"`
	Result                 string   `json:"result"`
	ChannelCode            string   `json:"-"`
}

type CreateDisbursementFromOpenApiResponse struct {
	UUID       string                  `json:"uuid"`
	MerchantID string                  `json:"merchantId"`
	Payouts    []PayoutObjectForCreate `json:"payouts"`
	Status     string                  `json:"status"`
	CreatedAt  time.Time               `json:"created"`
	UpdatedAt  time.Time               `json:"updated"`
}

type RetryDisbursementFromOpenApiResponse struct {
	UUID       string                 `json:"uuid"`
	MerchantID string                 `json:"merchantId"`
	Payouts    []PayoutObjectForRetry `json:"payouts"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created"`
	UpdatedAt  time.Time              `json:"updated"`
}

type GetBulkDisbursementForOpenApiByIDResponse struct {
	UUID          string             `json:"uuid"`
	MerchantID    string             `json:"merchantId"`
	PayoutResults PayoutResultObject `json:"payoutResults"`
	Payouts       []PayoutObject     `json:"payouts"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"created"`
	UpdatedAt     time.Time          `json:"updated"`
}

type GetBulkDisbursementForOpenApiByReferenceIDResponse struct {
	UUID          string             `json:"uuid"`
	MerchantID    string             `json:"merchantId"`
	PayoutResults PayoutResultObject `json:"payoutResults"`
	Payouts       PayoutObject       `json:"payouts"`
}

type ApprovalActionsResponse struct {
	FileRejected             string               `json:"fileRejected"`
	BeneficiaryLimitExceeded []ApprovalValidation `json:"beneficiaryLimitExceeded,omitempty"`
}

type GetDisbursementReceiptResponse struct {
	ReceiptURL string `json:"receiptUrl"`
}

type ExportDisbursementListResponse struct {
	Url string `json:"url"`
}

type DisbursementWithTransactionResponse struct {
	DisbursementWithTransaction

	FailedReason    *string                             `json:"failedReason"`
	RejectReason    *string                             `json:"rejectReason"`
	StatusHistories []DisbursementStatusHistoryResponse `json:"statusHistory,omitempty"`
}

type DisbursementStatusHistoryResponse struct {
	Status         string     `json:"status"`
	Label          string     `json:"label"`
	Description    string     `json:"description,omitempty"`
	Recommendation string     `json:"recommendation,omitempty"`
	Timestamp      *time.Time `json:"timestamp,omitempty"`
}

func (d *DisbursementWithTransaction) DisbursementWithTransactionToResponse() *DisbursementWithTransactionResponse {
	resp := &DisbursementWithTransactionResponse{DisbursementWithTransaction: *d}

	if d.Status == constant.DisbursementStatusRejected {
		resp.RejectReason = d.ReasonDescription
	}

	disbursementReasonType := ""
	transactionReasonType := ""
	if d.ReasonType != nil {
		disbursementReasonType = *d.ReasonType
	}
	if d.TransactionReasonType != nil {
		transactionReasonType = *d.TransactionReasonType
	}

	if disbursementReasonType == constant.DisbursementReasonTypeInsufficientBalance {
		resp.FailedReason = d.ReasonDescription
	} else if transactionReasonType == constant.ReasonTypeBeneficiaryAccountReason {
		resp.FailedReason = d.TransactionReasonDescription
	}

	if resp.Metadata.Valid {
		_ = json.Unmarshal(resp.Metadata.JSONText, &resp.MetadataObj)
	}

	if len(d.StatusHistories) > 0 {
		labelIndexMap := make(map[string]int)
		var statusHistoriesResp []DisbursementStatusHistoryResponse

		for _, history := range d.StatusHistories {
			statusHistoryResp := DisbursementStatusHistoryResponse{
				Status:    history.Status,
				Timestamp: &history.CreatedAt,
			}
			if history.MetadataObj != nil {
				statusHistoryResp.Label = history.MetadataObj.Label
				statusHistoryResp.Description = history.MetadataObj.Description
				statusHistoryResp.Recommendation = history.MetadataObj.Recommendation
			}

			if statusHistoryResp.Label != "" {
				if existingIndex, exists := labelIndexMap[statusHistoryResp.Label]; exists {
					statusHistoriesResp[existingIndex] = statusHistoryResp
				} else {
					labelIndexMap[statusHistoryResp.Label] = len(statusHistoriesResp)
					statusHistoriesResp = append(statusHistoriesResp, statusHistoryResp)
				}
			} else {
				statusHistoriesResp = append(statusHistoriesResp, statusHistoryResp)
			}
		}

		// Add dummy future responses based on current status
		if len(statusHistoriesResp) > 0 {
			lastStatus := statusHistoriesResp[len(statusHistoriesResp)-1].Status

			switch lastStatus {
			case constant.DisbursementStatusHistoryWaiting, constant.DisbursementStatusHistoryWaitingForTopUp:
				// Add PROCESSING dummy response
				processingLabels := constant.DisbursementStatusHistoryLabelsAndDescriptions[constant.DisbursementStatusHistoryProcessing]
				statusHistoriesResp = append(statusHistoriesResp, DisbursementStatusHistoryResponse{
					Status:      constant.DisbursementStatusHistoryProcessing,
					Label:       processingLabels["label"],
					Description: processingLabels["description"],
				})

				// Add SUCCESS dummy response
				successLabels := constant.DisbursementStatusHistoryLabelsAndDescriptions[constant.DisbursementStatusHistorySuccess]
				statusHistoriesResp = append(statusHistoriesResp, DisbursementStatusHistoryResponse{
					Status:      constant.DisbursementStatusHistorySuccess,
					Label:       successLabels["label"],
					Description: successLabels["description"],
				})

			case constant.DisbursementStatusHistoryProcessing:
				// Add SUCCESS dummy response
				successLabels := constant.DisbursementStatusHistoryLabelsAndDescriptions[constant.DisbursementStatusHistorySuccess]
				statusHistoriesResp = append(statusHistoriesResp, DisbursementStatusHistoryResponse{
					Status:      constant.DisbursementStatusHistorySuccess,
					Label:       successLabels["label"],
					Description: successLabels["description"],
				})
			}
		}

		resp.StatusHistories = statusHistoriesResp
	}

	return resp
}

type TransactionConfigResp struct {
	*TransactionConfig
	FeeDetail feeModel.FeeResponse `json:"feeDetail"` // Deprecated
}

type ReversalTransactionResp struct {
	Id             string  `json:"id"`
	DisbursementId string  `json:"disbursementId"`
	ReversalAmount float64 `json:"reversalAmount"`
}

type BulkCreateResponse struct {
	UUID        string    `json:"uuid"`
	MerchantID  string    `json:"merchantId"`
	File        string    `json:"file"`
	FileFailed  *string   `json:"fileFailed"`
	Status      string    `json:"status"`
	CreatedBy   *string   `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TotalData   int       `json:"totalData"`
	TotalAmount float64   `json:"totalAmount"`
}

type DailyTransactionLimitResponse struct {
	Limit     *float64 `json:"limit" redis:"limit" db:"limit"`
	Processed float64  `json:"processed" redis:"processed" db:"processed"`
	Remaining float64  `json:"remaining" redis:"-" db:"-"`
}

type ChangeDisbursementTransactionStatusResponse struct {
	DisbursementID string `json:"disbursementId"`
	Updated        bool   `json:"updated"`
	Reason         string `json:"reason"`
}

type CheckDisbursementTransactionStatusResponse struct {
	DisbursementID       string                                               `json:"disbursementId"`
	TransactionStatus    string                                               `json:"transactionStatus"`
	TransactionUpdatedAt time.Time                                            `json:"transactionUpdatedAt"`
	ProcessorData        *snapCoreModelBT.BankTransferCheckStatusResponseData `json:"processorData"`

	Error        bool   `json:"error,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (t DailyTransactionLimitResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *DailyTransactionLimitResponse) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, t)
}

type AfterPayoutCutOffTimeSummary struct {
	Total     int64                              `json:"total"`
	Amount    float64                            `json:"amount"`
	Banks     []AfterPayoutCutOffTimeBankSummary `json:"banks"`
	Merchants []string                           `json:"merchants,omitempty"`
	Info      string                             `json:"info"`
}

type PartnerWindowCutOffReportRequest struct {
	PartnerCode   string
	PartnerName   string
	WindowStartAt time.Time
	WindowEndAt   time.Time
	ExternalIDs   []string
}

type AfterPayoutCutOffTimeBankSummary struct {
	Name   string  `json:"name" db:"name"`
	Total  int64   `json:"total" db:"total"`
	Amount float64 `json:"amount" db:"amount"`
}

type RetryInquireDisbuesementSummary struct {
	Total           int64   `json:"total"`
	Amount          float64 `json:"amount"`
	TotalSucceeded  int64   `json:"totalSucceeded"`
	AmountSucceeded float64 `json:"amountSucceeded"`
	TotalFailed     int64   `json:"totalFailed"`
	AmountFailed    float64 `json:"amountFailed"`
}

type CancelDisbursementResponse struct {
	Total        int      `json:"total"`
	CancelledIds []string `json:"cancelledIds"`
}

type CRMPayoutStatusResponse struct {
	Code string                       `json:"code"`
	Data *CRMPayoutStatusResponseData `json:"data"`
}

type CRMPayoutStatusResponseData struct {
	DisbursementUUID   string               `json:"disbursementUuid"`
	ReferenceID        string               `json:"referenceId"`
	Status             string               `json:"status"`
	ApprovalStatus     string               `json:"approvalStatus"`
	Amount             string               `json:"amount"`
	BeneficiaryAccount string               `json:"beneficiaryAccount"`
	BeneficiaryName    string               `json:"beneficiaryName"`
	BeneficiaryBank    string               `json:"beneficiaryBank"`
	TransactionDate    string               `json:"transactionDate"`
	CreatedAt          string               `json:"createdAt"`
	UpdatedAt          string               `json:"updatedAt"`
	TransferLogs       []RoutingHistoryItem `json:"transferLogs"`
}

type RoutingHistoryItem struct {
	Order       int    `json:"order"`
	BankName    string `json:"bankName"`
	Status      string `json:"status"`
	ResponseMsg string `json:"responseMessage"`
	Timestamp   string `json:"timestamp"`
}

type LatestPayoutStatusResponse struct {
	RealTimeStatus   string `json:"realTimeStatus"`
	BankResponse     string `json:"bankResponse"`
	LastCheckedAt    string `json:"lastCheckedAt"`
	IsStatusFinal    bool   `json:"isStatusFinal"`
	NextCheckAllowed bool   `json:"nextCheckAllowed"`
}

type CRMBatchPayoutStatusResponse struct {
	Code string                  `json:"code"`
	Data []CRMPayoutStatusResult `json:"data"`
}

type CRMPayoutStatusResult struct {
	ReferenceID string                       `json:"referenceId"`
	Success     bool                         `json:"success"`
	Data        *CRMPayoutStatusResponseData `json:"data,omitempty"`
	Error       *CRMPayoutStatusError        `json:"error,omitempty"`
}

type CRMPayoutStatusError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CRMBatchSummary struct {
	Total    int `json:"total"`
	Success  int `json:"success"`
	Failed   int `json:"failed"`
	NotFound int `json:"notFound"`
}
