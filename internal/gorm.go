package internal

func IsValidFieldName(field string) bool {
	if len(field) == 0 {
		return false
	}

	for _, char := range field {
		if !isValidChar(char) {
			return false
		}
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
