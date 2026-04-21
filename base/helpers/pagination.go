package helpers

// GetOffset calculates the offset for SQL queries based on the current page and the number of items per page (limit).
// It's used for implementing pagination in database queries.
func GetOffset(page int, limit int) int {
	// Subtracting 1 from page number because pages usually start at 1, but offsets in SQL start at 0.
	// Then, multiply by the limit to get the offset.
	return (page - 1) * limit
}
