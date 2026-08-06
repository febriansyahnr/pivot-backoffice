package rolePermissionModel

type RolePermission struct {
	RoleID       string `json:"roleId" db:"role_id"`
	PermissionID string `json:"permissionId" db:"permission_id"`
}
