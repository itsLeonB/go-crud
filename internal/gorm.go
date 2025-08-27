package internal

// IsValidFieldName returns true if field contains only ASCII letters, digits,
// underscores, and single dots used as separators (e.g., "table.column").
// It rejects empty strings, leading/trailing dots, and consecutive dots.
func IsValidFieldName(field string) bool {
	if len(field) == 0 {
		return false
	}
	var prevDot bool
	for i, char := range field {
		if char == '.' {
			// No leading dot and no consecutive dots.
			if i == 0 || prevDot {
				return false
			}
			prevDot = true
			continue
		}
		prevDot = false
		if !isValidChar(char) {
			return false
		}
	}
	// No trailing dot.
	if prevDot {
		return false
	}
	return true
}

func isValidChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' ||
		char == '.'
}
