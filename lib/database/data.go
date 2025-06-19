package database

// Define ErrType as a string type
type ErrType string

// Define the error constant
const (
	ErrUnset    ErrType = ""
	ErrNotFound ErrType = "NOT_FOUND"
)
