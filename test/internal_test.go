package gocrud_test

import (
	"testing"

	"github.com/itsLeonB/go-crud/internal"
	"github.com/stretchr/testify/assert"
)

func TestIsValidFieldName(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      bool
	}{
		// Valid cases
		{"valid simple field", "name", true},
		{"valid field with underscore", "created_at", true},
		{"valid field with numbers", "field123", true},
		{"valid field with dot notation", "users.name", true},
		{"valid field with mixed case", "firstName", true},
		{"field starting with number", "123field", true},
		{"field starting with underscore", "_field", true},
		{"field with only underscores", "___", true},
		{"field with only numbers", "123", true},
		{"valid complex field", "table_name.field_123", true},
		{"valid multiple segments", "schema.table.column", true},
		
		// Invalid cases - empty and basic validation
		{"empty field name", "", false},
		{"field with space", "field name", false},
		{"field with special characters", "field@name", false},
		{"field with SQL injection attempt", "name'; DROP TABLE users; --", false},
		{"field with parentheses", "COUNT(*)", false},
		{"field with quotes", "name'test", false},
		{"unicode characters", "naméñ", false},
		
		// Invalid cases - dot validation (new strict rules)
		{"field starting with dot", ".field", false},
		{"field ending with dot", "field.", false},
		{"field with consecutive dots", "table..column", false},
		{"field with multiple consecutive dots", "table...column", false},
		{"only dots", "...", false},
		{"single dot", ".", false},
		{"leading dot with valid content", ".table.column", false},
		{"trailing dot with valid content", "table.column.", false},
		{"consecutive dots in middle", "table..column.field", false},
		
		// Invalid cases - other special characters
		{"field with hash", "name#test", false},
		{"field with dollar sign", "name$test", false},
		{"field with percent", "name%test", false},
		{"field with ampersand", "name&test", false},
		{"field with plus", "name+test", false},
		{"field with minus", "name-test", false},
		{"field with slash", "name/test", false},
		{"field with backslash", "name\\test", false},
		{"field with pipe", "name|test", false},
		{"field with brackets", "name[test]", false},
		{"field with braces", "name{test}", false},
		{"field with semicolon", "name;test", false},
		{"field with colon", "name:test", false},
		{"field with comma", "name,test", false},
		{"field with less than", "name<test", false},
		{"field with greater than", "name>test", false},
		{"field with question mark", "name?test", false},
		{"field with exclamation", "name!test", false},
		{"field with tilde", "name~test", false},
		{"tab character", "name\ttest", false},
		{"newline character", "name\ntest", false},
		{"carriage return", "name\rtest", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.IsValidFieldName(tt.fieldName)
			assert.Equal(t, tt.want, got, "IsValidFieldName(%q) should return %v", tt.fieldName, tt.want)
		})
	}
}

func TestIsValidFieldName_EdgeCases(t *testing.T) {
	// Test with very long string
	longField := make([]byte, 1000)
	for i := range longField {
		longField[i] = 'a'
	}
	
	assert.True(t, internal.IsValidFieldName(string(longField)), "IsValidFieldName should handle very long valid strings")

	// Test with string containing only valid characters (no dots at start/end)
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	assert.True(t, internal.IsValidFieldName(validChars), "IsValidFieldName should accept string with all valid characters")

	// Test valid dot usage
	validDotCases := []struct {
		name  string
		field string
		want  bool
	}{
		{"single dot separator", "table.column", true},
		{"multiple dot separators", "schema.table.column", true},
		{"dot with underscores", "table_name.column_name", true},
		{"dot with numbers", "table1.column2", true},
	}

	for _, tt := range validDotCases {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.IsValidFieldName(tt.field)
			assert.Equal(t, tt.want, got, "IsValidFieldName(%q) should return %v", tt.field, tt.want)
		})
	}

	// Test invalid dot usage
	invalidDotCases := []struct {
		name  string
		field string
		want  bool
	}{
		{"leading dot", ".table", false},
		{"trailing dot", "table.", false},
		{"consecutive dots", "table..column", false},
		{"triple dots", "table...column", false},
		{"leading and trailing dots", ".table.", false},
		{"only dots", "...", false},
		{"single dot only", ".", false},
		{"dot at start of complex field", ".schema.table.column", false},
		{"dot at end of complex field", "schema.table.column.", false},
		{"consecutive dots in complex field", "schema..table.column", false},
	}

	for _, tt := range invalidDotCases {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.IsValidFieldName(tt.field)
			assert.Equal(t, tt.want, got, "IsValidFieldName(%q) should return %v", tt.field, tt.want)
		})
	}

	// Test single character cases
	singleCharTests := []struct {
		char string
		want bool
	}{
		{"a", true},
		{"Z", true},
		{"0", true},
		{"_", true},
		{".", false}, // Single dot is now invalid
		{" ", false},
		{"@", false},
		{"'", false},
		{";", false},
	}

	for _, tt := range singleCharTests {
		t.Run("single_char_"+tt.char, func(t *testing.T) {
			got := internal.IsValidFieldName(tt.char)
			assert.Equal(t, tt.want, got, "IsValidFieldName(%q) should return %v", tt.char, tt.want)
		})
	}
}

func BenchmarkIsValidFieldName(b *testing.B) {
	testCases := []string{
		"simple_field",
		"table.column",
		"schema.table.column",
		"very_long_field_name_with_many_characters_and_numbers_123456789",
		"invalid field with spaces",
		"field'; DROP TABLE users; --",
		".invalid_leading_dot",
		"invalid_trailing_dot.",
		"invalid..consecutive.dots",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				internal.IsValidFieldName(tc)
			}
		})
	}
}

// Test concurrent access to ensure thread safety
func TestIsValidFieldName_Concurrent(t *testing.T) {
	testCases := []struct {
		field    string
		expected bool
	}{
		{"concurrent_test_field", true},
		{"table.column", true},
		{".invalid", false},
		{"invalid.", false},
		{"invalid..dots", false},
	}
	
	// Run multiple goroutines concurrently
	done := make(chan bool, 500) // 100 goroutines * 5 test cases
	
	for i := 0; i < 100; i++ {
		go func() {
			for _, tc := range testCases {
				result := internal.IsValidFieldName(tc.field)
				assert.Equal(t, tc.expected, result, "IsValidFieldName should return %v for %q", tc.expected, tc.field)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
}

// Test specific SQL injection patterns
func TestIsValidFieldName_SQLInjectionPrevention(t *testing.T) {
	sqlInjectionAttempts := []string{
		"name'; DROP TABLE users; --",
		"name' OR '1'='1",
		"name'; DELETE FROM users; --",
		"name' UNION SELECT * FROM passwords --",
		"name'; INSERT INTO users VALUES ('hacker'); --",
		"name' AND 1=1 --",
		"name'; EXEC xp_cmdshell('dir'); --",
		"name' OR 1=1#",
		"name'; UPDATE users SET admin=1; --",
		"name' OR 'x'='x",
	}

	for _, attempt := range sqlInjectionAttempts {
		t.Run("sql_injection_"+attempt, func(t *testing.T) {
			result := internal.IsValidFieldName(attempt)
			assert.False(t, result, "IsValidFieldName should reject SQL injection attempt: %q", attempt)
		})
	}
}

// Test edge cases with different character encodings
func TestIsValidFieldName_CharacterEncoding(t *testing.T) {
	encodingTests := []struct {
		name  string
		field string
		want  bool
	}{
		{"ascii_only", "valid_field_123", true},
		{"utf8_unicode", "field_ñame", false},
		{"utf8_emoji", "field_😀", false},
		{"utf8_chinese", "field_中文", false},
		{"utf8_arabic", "field_العربية", false},
		{"control_chars", "field\x00name", false},
		{"high_ascii", "field\x80name", false},
	}

	for _, tt := range encodingTests {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.IsValidFieldName(tt.field)
			assert.Equal(t, tt.want, got, "IsValidFieldName(%q) should return %v", tt.field, tt.want)
		})
	}
}
