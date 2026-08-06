package constant

var PredefinedRoles = []string{RoleAdmin, RoleMaker, RoleApprover, RoleDeveloper, RoleOperation, RoleCrossborderOperator, RolePlatform, RoleVccOperator}

const (
	RoleAdmin               = "ADMIN"
	RoleMaker               = "MAKER"
	RoleApprover            = "APPROVER"
	RoleDeveloper           = "DEVELOPER"
	RoleOperation           = "OPERATION"
	RoleCrossborderOperator = "CROSSBORDER_OPERATOR"
	RolePlatform            = "PLATFORM"
	RoleVccOperator         = "VCC_OPERATOR"

	RoleTypeDefault = "DEFAULT"
	RoleTypeCustom  = "CUSTOM"

	MaxRolesPerMerchant uint64 = 10
)
