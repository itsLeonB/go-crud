package gocrud_test

import (
	"testing"

	"github.com/itsLeonB/go-crud/lib"
	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
	}{
		{
			name:     "ContextKeyGormTx value",
			constant: string(lib.ContextKeyGormTx),
			expected: "go-crud.gormTx",
		},
		{
			name:     "MsgTransactionError value",
			constant: lib.MsgTransactionError,
			expected: "error processing transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant, "Constant %s should have expected value", tt.name)
		})
	}
}

func TestContextKeyGormTx_Type(t *testing.T) {
	// Verify that ContextKeyGormTx is of the correct type
	// The constant is defined as txKey type, not string directly
	keyStr := string(lib.ContextKeyGormTx)
	assert.Equal(t, "go-crud.gormTx", keyStr, "ContextKeyGormTx string value should be correct")
}

func TestMsgTransactionError_Type(t *testing.T) {
	// Verify that MsgTransactionError is a string
	assert.IsType(t, "", lib.MsgTransactionError, "MsgTransactionError should be a string")
	assert.Equal(t, "error processing transaction", lib.MsgTransactionError, "MsgTransactionError should have correct value")
}

func TestConstants_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, lib.ContextKeyGormTx, "ContextKeyGormTx should not be empty")
	assert.NotEmpty(t, lib.MsgTransactionError, "MsgTransactionError should not be empty")
}

func TestConstants_Uniqueness(t *testing.T) {
	constants := map[string]interface{}{
		"ContextKeyGormTx":    lib.ContextKeyGormTx,
		"MsgTransactionError": lib.MsgTransactionError,
	}

	// Check that all constants have different values
	values := make(map[interface{}]string)
	for name, value := range constants {
		if existingName, exists := values[value]; exists {
			t.Errorf("Constants %s and %s have the same value: %v", name, existingName, value)
		}
		values[value] = name
	}
}

func TestContextKeyGormTx_Usage(t *testing.T) {
	// Should be usable as a context key (comparable type)
	key1 := lib.ContextKeyGormTx
	key2 := lib.ContextKeyGormTx

	assert.Equal(t, key1, key2, "ContextKeyGormTx should be comparable and equal to itself")

	// Should be different from other string values
	otherKey := "some.other.key"
	assert.NotEqual(t, string(key1), otherKey, "ContextKeyGormTx should be unique")
}

func TestMsgTransactionError_Usage(t *testing.T) {
	// Test that the message is suitable for error messages
	msg := lib.MsgTransactionError

	assert.GreaterOrEqual(t, len(msg), 5, "MsgTransactionError should be a meaningful message")

	// Should not contain special characters that might break error handling
	for _, char := range msg {
		assert.True(t, char >= 32 && char <= 126, "MsgTransactionError should not contain non-printable character: %d", char)
	}
}
