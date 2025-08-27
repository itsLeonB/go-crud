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
		{"valid simple field", "name", true},
		{"valid field with underscore", "created_at", true},
		{"valid field with numbers", "field123", true},
		{"valid field with dot notation", "users.name", true},
		{"valid field with mixed case", "firstName", true},
		{"empty field name", "", false},
		{"field with space", "field name", false},
		{"field with special characters", "field@name", false},
		{"field with SQL injection attempt", "name'; DROP TABLE users; --", false},
		{"field with parentheses", "COUNT(*)", false},
		{"field with quotes", "name'test", false},
		{"field starting with number", "123field", true},
		{"field starting with underscore", "_field", true},
		{"field with only underscores", "___", true},
		{"unicode characters", "naméñ", false},
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

	// Test with string containing only valid characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_."
	assert.True(t, internal.IsValidFieldName(validChars), "IsValidFieldName should accept string with all valid characters")

	// Test single character cases
	singleCharTests := []struct {
		char string
		want bool
	}{
		{"a", true},
		{"Z", true},
		{"0", true},
		{"_", true},
		{".", true},
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
		"very_long_field_name_with_many_characters_and_numbers_123456789",
		"invalid field with spaces",
		"field'; DROP TABLE users; --",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				internal.IsValidFieldName(tc)
			}
		})
	}
}
