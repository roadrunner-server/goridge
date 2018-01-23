package goridge

// min provides simple uint64 comparison
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
