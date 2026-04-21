package user

import (
	dto "template-go/base/base"
	"template-go/data/enum"
)

type FilterUser struct {
	dto.PaginateIn
	Role enum.Role
}
