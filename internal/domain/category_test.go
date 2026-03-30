package domain_test

import (
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCategory_IsSystem(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
		want   bool
	}{
		{"user ID 0 is system category", 0, true},
		{"positive user ID is user category", 1, false},
		{"large user ID is user category", 123456789, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &domain.Category{UserID: tt.userID}
			assert.Equal(t, tt.want, c.IsSystem())
		})
	}
}
