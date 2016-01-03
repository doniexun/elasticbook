package utils

// ContainsString returns true iff slice contains element
func ContainsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for i, e := range slice {
		if e == element {
			return i
		}
	}
	return -1
}
