package menuModel

type GetAllFilterRequest struct {
	RoleID      string
	ExcludeHome bool // When true, excludes Home menu from results (used for role management UI)
}
