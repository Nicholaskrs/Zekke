package api

const (
	MaxLimit     = 20
	DefaultLimit = 10
)

// PaginationMeta used for pagination. Will be set into nil if there is no pagination in the BaseResponse.Data.
type PaginationMeta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	TotalRows int `json:"totalRows"`
}

type PaginateRequest struct {
	Page  int `form:"page" binding:"required,min=1"`
	Limit int `form:"limit" binding:"required,min=1,max=100"`
}

func (p *PaginateRequest) GetLimit() int {
	if p.Limit > MaxLimit {
		return MaxLimit
	}
	if p.Limit < 1 {
		return DefaultLimit
	}
	return p.Limit
}
