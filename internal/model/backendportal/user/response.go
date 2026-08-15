package user

import (
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/tnc"
)

type ValidateInvitationResponse struct {
	UserID       string                `json:"userId" redis:"userId"`
	UserName     string                `json:"userName" redis:"userName"`
	Email        string                `json:"email" redis:"email"`
	MerchantName string                `json:"merchantName" redis:"merchantName"`
	MerchantID   string                `json:"merchantId" redis:"merchantId"`
	TNCMetadata  *tnc.TNCSigningStatus `json:"tncMetadata" redis:"-"`
}

func (v ValidateInvitationResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

func (v *ValidateInvitationResponse) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, v)
}

type UpdateStatusResp struct {
	Updated bool `json:"updated"`
}

type InvitationURLResp struct {
	URL string `json:"url"`
}
