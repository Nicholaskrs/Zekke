package dto

type BaseOut struct {
	Success      bool
	ErrorMessage string
	ErrorCode    int
}

type PaginateIn struct {
	Limit int
	Page  int
}

func (p *PaginateIn) GetOffset() int {
	// Subtracting 1 from page number because pages usually start at 1, but offsets in SQL start at 0.
	// Then, multiply by the limit to get the offset.
	return (p.Page - 1) * p.Limit
}
