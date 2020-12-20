package shared_memory //nolint:golint,stylecheck

// SharedMemory interface represents shared memory segment
type SharedMemory interface {
	// Read specified amount of data (len)
	Read(data []byte) error
	// Write data to the shared memory segment
	Write(data []byte)
	// Detach from the segment
	Detach() error
	// Clear the shared memory segment
	// By semantic similar to std::memset(..., 0, ...)
	Clear()
}
