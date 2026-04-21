package constants

type MethodType string

const (
	MethodFetch  MethodType = "fetch"
	MethodGet    MethodType = "get"
	MethodCreate MethodType = "create"
	MethodUpdate MethodType = "update"
	MethodDelete MethodType = "delete"

	MethodNull MethodType = ""
)
