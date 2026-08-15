package merchant

type SubMerchantAdminRequest struct {
	MerchantId string `json:"-" validate:"required,uuid"`
	Email      string `json:"email" validate:"required,email"`
	Name       string `json:"name" validate:"required"`
	Invitation bool   `json:"invitation" validate:"-"`
}

type ResendInvitationRequest struct {
	Email string `json:"email" validate:"required,email"`

	MerchantId       string `json:"-" validate:"required,uuid"`
	ParentMerchantId string `json:"-" validate:"required,uuid"`
}
