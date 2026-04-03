package validator_test

import (
	"testing"

	"notification-service/pkg/validator"

	"github.com/stretchr/testify/assert"
)

func TestEmailValidator_IsValid_WithValidEmails_ReturnsTrue(t *testing.T) {
	v := validator.NewEmailValidator()

	validEmails := []string{
		"user@example.com",
		"test.user@domain.org",
		"name+tag@company.co",
		"user123@test.io",
		"a@b.cd",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			assert.True(t, v.IsValid(email), "email %q should be valid", email)
		})
	}
}

func TestEmailValidator_IsValid_WithInvalidEmails_ReturnsFalse(t *testing.T) {
	v := validator.NewEmailValidator()

	invalidEmails := []string{
		"",
		"not-an-email",
		"@domain.com",
		"user@",
		"user@.com",
		"user space@domain.com",
		"user@domain",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			assert.False(t, v.IsValid(email), "email %q should be invalid", email)
		})
	}
}

func TestIsValidEmail_WithValidEmail_ReturnsTrue(t *testing.T) {
	assert.True(t, validator.IsValidEmail("test@example.com"))
}

func TestIsValidEmail_WithInvalidEmail_ReturnsFalse(t *testing.T) {
	assert.False(t, validator.IsValidEmail("invalid"))
}
