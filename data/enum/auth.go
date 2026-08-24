package enum

type Role string

const (
	Admin  Role = "Admin"
	Member Role = "Member"
)

func (s Role) String() string {
	return string(s)
}

var SliceRole = []string{
	Admin.String(),
	Member.String(),
}
