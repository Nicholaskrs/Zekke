package permission

import (
	dto "template-go/base/base"
)

type FilterPrivilegeGroup struct {
	dto.PaginateIn
	CodeStartWith string
}

type FilterPrivilegeGroupPrivilege struct {
	dto.PaginateIn
	PrivilegeGroup string
}
