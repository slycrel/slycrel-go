package io

import "time"

// Now returns the current time. Extracted to a function for future testability.
func Now() time.Time {
	return time.Now()
}
