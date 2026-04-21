package enum

type Role string

const (
	National    Role = "National"
	AreaManager Role = "Area Manager"
	Distributor Role = "Distributor"
	Operator    Role = "Operator"
	Sales       Role = "Sales"
)

func (s Role) String() string {
	return string(s)
}

var SliceRole = []string{
	National.String(),
	AreaManager.String(),
	Distributor.String(),
	Operator.String(),
	Sales.String(),
}
