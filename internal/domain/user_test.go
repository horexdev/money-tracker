package domain_test

import (
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestUser_DisplayName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		username  string
		want      string
	}{
		{"first name takes priority", "Alice", "alice_tg", "Alice"},
		{"username when no first name", "", "alice_tg", "@alice_tg"},
		{"fallback to User when both empty", "", "", "User"},
		{"only first name set", "Bob", "", "Bob"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &domain.User{FirstName: tt.firstName, Username: tt.username}
			assert.Equal(t, tt.want, u.DisplayName())
		})
	}
}

func TestValidTheme(t *testing.T) {
	tests := []struct {
		theme string
		want  bool
	}{
		{"system", true},
		{"light", true},
		{"dark", true},
		{"", false},
		{"auto", false},
		{"Dark", false},
		{"SYSTEM", false},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.ValidTheme(tt.theme))
		})
	}
}

func TestValidLanguage(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"en", true},
		{"ru", true},
		{"uk", true},
		{"be", true},
		{"kk", true},
		{"uz", true},
		{"es", true},
		{"de", true},
		{"it", true},
		{"fr", true},
		{"pt", true},
		{"nl", true},
		{"ar", true},
		{"tr", true},
		{"ko", true},
		{"ms", true},
		{"id", true},
		{"zh", false},
		{"xx", false},
		{"", false},
		{"EN", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.ValidLanguage(tt.code))
		})
	}
}
