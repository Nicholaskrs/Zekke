package dto

type BaseOut struct {
	Success      bool
	ErrorMessage string
	ErrorCode    int
}

type PaginateOut struct {
	TotalRows int64
}

type PaginateIn struct {
	Limit int
	Page  int
}
