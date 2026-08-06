package roleMenuPermissionModel

type MenuPermissionFromFileRequest struct {
	Slug        string                          `json:"slug"`
	Name        string                          `json:"name"`
	Icon        string                          `json:"icon"`
	Path        string                          `json:"path"`
	Type        string                          `json:"type,omitempty"`
	Children    []MenuPermissionFromFileRequest `json:"children,omitempty"`
	Permissions []Permissions                   `json:"permissions"`
}

type Permissions struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Group       string   `json:"group"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}
