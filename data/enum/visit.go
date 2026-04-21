package enum

type VisitFetchType string

// Visit Type used to fetch visitation with filter for today visitation or all history
const (
	// Type history will fetch only today's visit.
	VisitFetchTypeToday VisitFetchType = "today"
	// Type history will fetch all historical data before current date.
	VisitFetchTypeHistory VisitFetchType = "history"
)

func (s VisitFetchType) String() string {
	return string(s)
}
