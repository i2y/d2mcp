package usecase

// ValidationError represents a validation error in the usecase layer.
type ValidationError struct {
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}
