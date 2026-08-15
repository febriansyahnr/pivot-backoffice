package permissionModel

type PermissionFromFileRequest struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Group       string   `json:"group"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}
