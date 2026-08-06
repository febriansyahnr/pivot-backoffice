package commonModel

type PaginationResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

type Meta struct {
	Page       int64 `json:"page" example:"1"`
	PerPage    int64 `json:"perPage" example:"10"`
	TotalItems int64 `json:"totalItems" example:"100"`
	TotalPages int64 `json:"totalPages" example:"10"`
}

func NewMeta(page, perPage, totalItems int64) *Meta {
	meta := Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalItems / perPage,
	}

	if totalItems%perPage != 0 {
		meta.TotalPages++
	}
	return &meta
}
