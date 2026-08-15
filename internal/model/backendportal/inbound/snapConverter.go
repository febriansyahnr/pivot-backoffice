package inboundModel

import (
	"fmt"
	"net/http"
	"time"

	"encoding/base64"
	"encoding/json"

	"github.com/jmoiron/sqlx/types"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	snap_signature "github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/signature"
)

const (
	SnapURLAccessToken  = "/snap/v1.0/access-token/b2b"
	SnapURLCreateVA     = "/snap/v1.0/transfer-va/create-va"
	SnapURLGenerateQRIS = "/snap/v1.0/qr/qr-mpm-generate"
)

type InboundSnapVersionResponse struct {
	ID                string             `json:"id"`
	ReferenceID       string             `json:"referenceId"`
	OriginID          string             `json:"originId"`
	TraceID           string             `json:"traceId"`
	IP                string             `json:"ip"`
	Method            string             `json:"method"`
	URL               string             `json:"url"`
	Headers           types.JSONText     `json:"headers"`
	Body              types.NullJSONText `json:"body"`
	StatusCode        int                `json:"statusCode"`
	ResponseTimeMs    float64            `json:"responseTimeMs"`
	ResponseBody      types.NullJSONText `json:"responseBody"`
	Metadata          types.NullJSONText `json:"metadata"`
	SnapCompatibility bool               `json:"snapCompatibility"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`

	Client  *Client `json:"-"`
	Feature string  `json:"feature"`
}

func (r *InboundResponse) ToSnapVersionResponse() *InboundSnapVersionResponse {
	resp := &InboundSnapVersionResponse{
		ID:                r.ID,
		ReferenceID:       r.ReferenceID,
		OriginID:          r.OriginID,
		TraceID:           r.TraceID,
		IP:                r.IP,
		Method:            r.Method,
		URL:               r.URL,
		Headers:           r.Headers,
		Body:              r.Body,
		StatusCode:        r.StatusCode,
		ResponseTimeMs:    r.ResponseTimeMs,
		ResponseBody:      r.ResponseBody,
		Metadata:          r.Metadata,
		SnapCompatibility: r.SnapCompatibility,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		Client:            r.Client,
		Feature:           r.Feature,
	}

	resp.convertToSnapVersion()

	return resp
}

func (r *InboundSnapVersionResponse) convertToSnapVersion() {
	var tempResponseBodyStruct struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(r.ResponseBody.JSONText, &tempResponseBodyStruct); err != nil {
		return
	}

	if r.StatusCode > 299 {
		r.convertErrToSnapVersion()
	}

	switch r.Feature {
	case constant.InboundFeaturePayment:
		paymentMethodType := ""
		createUnifiedPaymentRequest := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{}
		if err := json.Unmarshal(r.Body.JSONText, createUnifiedPaymentRequest); err != nil {
			return
		}
		confirmUnifiedPaymentRequest := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{}
		if err := json.Unmarshal(r.Body.JSONText, confirmUnifiedPaymentRequest); err != nil {
			return
		}

		if createUnifiedPaymentRequest.PaymentMethod != nil {
			paymentMethodType = createUnifiedPaymentRequest.PaymentMethod.Type
		} else if confirmUnifiedPaymentRequest.PaymentMethod != nil {
			paymentMethodType = confirmUnifiedPaymentRequest.PaymentMethod.Type
		}

		if paymentMethodType == constant.UnifiedPaymentMethodVA {
			r.URL = SnapURLCreateVA
		} else if paymentMethodType == constant.UnifiedPaymentMethodQris {
			r.URL = SnapURLGenerateQRIS
		}

		existingHeaders := map[string][]string{}
		_ = json.Unmarshal(r.Headers, &existingHeaders)
		r.Headers, _ = json.Marshal(map[string][]string{
			"Content-Type":  existingHeaders["Content-Type"],
			"Authorization": existingHeaders["Authorization"],
			"X-TIMESTAMP":   {r.CreatedAt.Format(util.SnapDateFormatLayout)},
			"X-SIGNATURE":   {generateDummyServiceSignature(http.MethodPost, r.URL, r.CreatedAt.Format(util.SnapDateFormatLayout), json.RawMessage(r.Body.JSONText))},
			"X-PARTNER-ID":  {r.ReferenceID},
			"X-EXTERNAL-ID": {r.OriginID}, // TODO: Confirm to product
			"CHANNEL-ID":    {"HARSYA"},
		})

		unifiedPaymentResponse := &unifiedPaymentModel.UnifiedPaymentSessionResponse{}
		if err := json.Unmarshal(tempResponseBodyStruct.Data, unifiedPaymentResponse); err != nil {
			return
		}

		if unifiedPaymentResponse.ChargeDetails == nil {
			return
		}

		if unifiedPaymentResponse.PaymentMethod.Type == constant.UnifiedPaymentMethodVA {
			r.convertPaymentVAToSnapVersion(unifiedPaymentResponse)
		} else if unifiedPaymentResponse.PaymentMethod.Type == constant.UnifiedPaymentMethodQris {
			r.convertPaymentQRToSnapVersion(unifiedPaymentResponse)
		}

	case constant.InboundFeatureAuth:
		r.convertAccessTokenB2BToSnapVersion()
	}
}

func (r *InboundSnapVersionResponse) convertPaymentVAToSnapVersion(unifiedPaymentResponse *unifiedPaymentModel.UnifiedPaymentSessionResponse) {
	if unifiedPaymentResponse.ChargeDetails[0].VirtualAccount == nil {
		return
	}

	bodyJson, _ := json.Marshal(map[string]any{
		"virtualAccountName":    unifiedPaymentResponse.ChargeDetails[0].VirtualAccount.VirtualAccountName,
		"virtualAccountTrxType": "CLOSED_DYNAMIC",
		"totalAmount": &commonModel.Amount{
			Currency: "IDR",
			Value:    fmt.Sprintf("%.2f", unifiedPaymentResponse.Amount.Value),
		},
		"expiredDate": unifiedPaymentResponse.ExpiryAt.Format(util.SnapDateFormatLayout),
		"additionalInfo": map[string]any{
			"referenceId": unifiedPaymentResponse.ClientReferenceID,
			"issuer":      unifiedPaymentResponse.ChargeDetails[0].VirtualAccount.Channel,
		},
	})
	r.Body = types.NullJSONText{
		JSONText: bodyJson,
		Valid:    true,
	}

	responseJson, _ := json.Marshal(map[string]any{
		"responseCode":    "2002700",
		"responseMessage": "Success",
		"virtualAccountData": map[string]any{
			"trxId":                 unifiedPaymentResponse.ClientReferenceID,
			"virtualAccountTrxType": "CLOSED_DYNAMIC",
			"virtualAccountNo":      unifiedPaymentResponse.ChargeDetails[0].VirtualAccount.VirtualAccountNumber,
			"virtualAccountName":    unifiedPaymentResponse.ChargeDetails[0].VirtualAccount.VirtualAccountName,
			"totalAmount": &commonModel.Amount{
				Currency: "IDR",
				Value:    fmt.Sprintf("%.2f", unifiedPaymentResponse.Amount.Value),
			},
			"expiredDate": unifiedPaymentResponse.ExpiryAt.Format(util.SnapDateFormatLayout),
		},
		"additionalInfo": map[string]any{
			"referenceId":   unifiedPaymentResponse.ClientReferenceID,
			"issuer":        unifiedPaymentResponse.ChargeDetails[0].VirtualAccount.Channel,
			"vaStatus":      "", // TODO: Confirm to product team
			"paymentStatus": unifiedPaymentResponse.ChargeDetails[0].Status,
		},
	})
	r.ResponseBody = types.NullJSONText{
		JSONText: responseJson,
		Valid:    true,
	}
}

func (r *InboundSnapVersionResponse) convertPaymentQRToSnapVersion(unifiedPaymentResponse *unifiedPaymentModel.UnifiedPaymentSessionResponse) {
	if unifiedPaymentResponse.ChargeDetails[0].Qr == nil {
		return
	}

	bodyJson, _ := json.Marshal(map[string]any{
		"amount": &commonModel.Amount{
			Currency: "IDR",
			Value:    fmt.Sprintf("%.2f", unifiedPaymentResponse.Amount.Value),
		},
		"additionalInfo": map[string]any{
			"qrType": "DYNAMIC",
		},
	})
	r.Body = types.NullJSONText{
		JSONText: bodyJson,
		Valid:    true,
	}

	responseJson, _ := json.Marshal(map[string]any{
		"responseCode":    "2004700",
		"responseMessage": "Success",
		"referenceNo":     unifiedPaymentResponse.ChargeDetails[0].Qr.RetrievalReferenceNumber,
		"merchantName":    unifiedPaymentResponse.ChargeDetails[0].StatementDescriptor,
		"qrContent":       unifiedPaymentResponse.ChargeDetails[0].Qr.QrContent,
		"qrUrl":           unifiedPaymentResponse.ChargeDetails[0].Qr.QrUrl,
		"qrImage":         base64.StdEncoding.EncodeToString([]byte(unifiedPaymentResponse.ChargeDetails[0].Qr.QrUrl)),
		"additionalInfo": map[string]any{
			"qrType":        "DYNAMIC",
			"qrStatus":      "", // TODO: Confirm to product
			"qrExpiredDate": unifiedPaymentResponse.ExpiryAt.Format(util.SnapDateFormatLayout),
			"paymentStatus": unifiedPaymentResponse.ChargeDetails[0].Status,
		},
	})
	r.ResponseBody = types.NullJSONText{
		JSONText: responseJson,
		Valid:    true,
	}
}

func (r *InboundSnapVersionResponse) convertAccessTokenB2BToSnapVersion() {
	r.URL = SnapURLAccessToken

	existingHeaders := map[string][]string{}
	_ = json.Unmarshal(r.Headers, &existingHeaders)
	r.Headers, _ = json.Marshal(map[string][]string{
		"Content-Type": existingHeaders["Content-Type"],
		"X-TIMESTAMP":  {r.CreatedAt.Format(util.SnapDateFormatLayout)},
		"X-CLIENT-KEY": {r.ReferenceID},
		"X-SIGNATURE":  {generateDummyAuthSignature(r.ReferenceID, r.CreatedAt.Format(util.SnapDateFormatLayout))},
	})

	bodyJson, _ := json.Marshal(map[string]any{
		"grantType":      "client_credentials",
		"additionalInfo": map[string]any{},
	})
	r.Body = types.NullJSONText{
		JSONText: bodyJson,
		Valid:    true,
	}

	var responseBodyStruct struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(r.ResponseBody.JSONText, &responseBodyStruct) != nil {
		return
	}
	existingResponseBody := map[string]any{}
	_ = json.Unmarshal(responseBodyStruct.Data, &existingResponseBody)

	responseJson, _ := json.Marshal(map[string]any{
		"responseCode":    "2007300",
		"responseMessage": "Successful",
		"accessToken":     existingResponseBody["accessToken"],
		"expiresIn":       existingResponseBody["expiresIn"],
		"tokenType":       existingResponseBody["tokenType"],
		"additionalInfo":  map[string]any{},
	})
	r.ResponseBody = types.NullJSONText{
		JSONText: responseJson,
		Valid:    true,
	}
}

func (r *InboundSnapVersionResponse) convertErrToSnapVersion() {
	// TODO: Map the error response SNAP
	responseJson, _ := json.Marshal(map[string]string{
		"responseCode":    fmt.Sprintf("%d27000", r.StatusCode),
		"responseMessage": "Error",
	})
	r.ResponseBody = types.NullJSONText{
		JSONText: responseJson,
		Valid:    true,
	}
}

func generateDummyAuthSignature(clientKey, timestamp string) string {
	tokenSignature := snap_signature.B2bTokenSignature{
		PrivateKey: util.GetMockPrivateKey(),
		Timestamp:  timestamp,
		ClientID:   clientKey,
	}

	signature, err := tokenSignature.Create()
	if err != nil {
		return ""
	}

	return signature
}

func generateDummyServiceSignature(httpMethod, url, timestamp string, body json.RawMessage) string {
	snapSignature, _ := snap_signature.NewTrxSignature(snap_signature.TrxSignature{
		HttpMethod:   httpMethod,
		URL:          url,
		AccessToken:  "Bearer dummy-access-token",
		ClientSecret: "dummy-client-secret",
		BodyPayload:  body,
		Timestamp:    timestamp,
	})
	signature := snapSignature.Create()
	if signature == nil {
		return ""
	}

	return *signature
}
