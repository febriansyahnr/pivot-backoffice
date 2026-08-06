package role

type RoleMenuResponse struct {
	ID    string                       `json:"id"`
	Name  string                       `json:"name"`
	Menus []RoleMenuPermissionResponse `json:"menus"`
}

type RoleMenuPermissionResponse struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permission"`
}
