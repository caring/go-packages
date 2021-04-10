package pointer

// String returns the value at the given string pointer, or a 0 value if it is nil
func String(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Int return the value at the given int point or a 0 value if it is nil
func Int(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// IntTo64 returns the value at the given int pointer cast to int64 or a zero value if it is nil
func IntTo64(i *int) int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

// Boolean returns boolean given boolean pointer
func Boolean(b *bool) bool {
	if b == nil {
		return false
	}
	return bool(*b)
}
