package constants

const (
	// MaxBatchSize defines the maximum number of rows per batch insert.
	// The batch insert limit is set to prevent potential errors and optimize performance.
	MaxBatchSize = 100
)
