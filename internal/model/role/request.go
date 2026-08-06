package role

type GetRoleFilterRequest struct {
	MerchantID string `json:"merchantId"`
}

type CreateRoleRequest struct {
	Name       string            `json:"name" validate:"required"`
	Menus      []RoleMenuRequest `json:"menus" validate:"min=1,unique=Slug,dive"`
	MerchantID string            `json:"-" validate:"required,uuid"`
}

type UpdateRoleRequest struct {
	Name       string            `json:"name" validate:"required"`
	Menus      []RoleMenuRequest `json:"menus" validate:"min=1,unique=Slug,dive"`
	MerchantID string            `json:"-" validate:"required,uuid"`
	RoleID     string            `json:"-" validate:"required,uuid"`
}

type RoleMenuRequest struct {
	Slug        string   `json:"slug" validate:"required"`
	Permissions []string `json:"permissions" validate:"min=1,unique,dive,required"`
}

type CRMUpdateDefaultRolePermissionsRequest struct {
	RoleSlug string            `json:"roleSlug" validate:"required,oneof=ADMIN MAKER APPROVER DEVELOPER OPERATION CROSSBORDER_OPERATOR PLATFORM"`
	Menus    []RoleMenuRequest `json:"menus" validate:"min=1,unique=Slug,dive"`
}
