package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"yflow/utils"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		expected     int
	}{
		{"Empty string", "", 10, 10},
		{"Valid integer", "42", 0, 42},
		{"Negative integer", "-42", 0, -42},
		{"Non-integer string", "abc", 5, 5},
		{"Mixed string", "123abc", 5, 5},
		{"Zero", "0", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ParseInt(tt.input, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseIntWithRange(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		min          int
		max          int
		expected     int
	}{
		{"Within range", "50", 0, 0, 100, 50},
		{"Below min", "10", 0, 20, 100, 20},
		{"Above max", "150", 0, 0, 100, 100},
		{"Invalid input, use default", "abc", 30, 0, 100, 30},
		{"Invalid input, default below min", "abc", 10, 20, 100, 20},
		{"Invalid input, default above max", "abc", 150, 0, 100, 100},
		{"Empty string", "", 50, 0, 100, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ParseIntWithRange(tt.input, tt.defaultValue, tt.min, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", false},
		{"Valid integer", "42", true},
		{"Negative integer", "-42", true},
		{"Zero", "0", true},
		{"Non-integer string", "abc", false},
		{"Mixed string", "123abc", false},
		{"Decimal number", "42.5", false},
		{"Only minus sign", "-", false},
		{"Space in number", "42 43", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidInteger(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		expected     int
	}{
		{"Valid positive", "42", 0, 42},
		{"Zero", "0", 5, 5},
		{"Negative", "-42", 5, 5},
		{"Invalid input", "abc", 5, 5},
		{"Empty string", "", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ParsePositiveInt(tt.input, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{"Within max length", "hello", 10, "hello"},
		{"Trim spaces", "  hello  ", 10, "hello"},
		{"Truncate to max length", "hello world", 5, "hello"},
		{"Truncate and trim", "  hello world  ", 5, "hello"},
		{"Empty string", "", 10, ""},
		{"Only spaces", "    ", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeString(tt.input, tt.maxLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		substrings []string
		expected   bool
	}{
		{"Contains one substring", "hello world", []string{"hello", "goodbye"}, true},
		{"Contains multiple substrings", "hello world", []string{"hello", "world"}, true},
		{"Contains no substrings", "hello world", []string{"goodbye", "universe"}, false},
		{"Empty input string", "", []string{"hello", "world"}, false},
		{"Empty substrings slice", "hello world", []string{}, false},
		{"Empty substring in slice", "hello world", []string{""}, true},
		{"Case sensitive match", "Hello World", []string{"hello"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ContainsAny(tt.input, tt.substrings)
			assert.Equal(t, tt.expected, result)
		})
	}
}
