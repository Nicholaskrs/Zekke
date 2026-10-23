package user

import dto "template-go/base/base"

type FilterUser struct {
	dto.PaginateIn
	Role      string
	IsDeleted bool
}
