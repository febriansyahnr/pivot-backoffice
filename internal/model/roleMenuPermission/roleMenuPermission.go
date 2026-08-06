package roleMenuPermissionModel

type RoleMenuPermission struct {
	RoleID       string `json:"roleId" db:"role_id"`
	MenuID       string `json:"menuId" db:"menu_id"`
	PermissionID string `json:"permissionId" db:"permission_id"`
}
