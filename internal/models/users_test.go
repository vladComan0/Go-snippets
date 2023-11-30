package models

import (
	"testing"

	"github.com/vladComan0/go-snippets/internal/assert"
)

func TestUserModelExists(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	tests := []struct {
		name   string
		userID int
		want   bool
	}{
		{
			name:   "Valid ID",
			userID: 1,
			want:   true,
		},
		{
			name:   "Non-existent ID",
			userID: 5,
			want:   false,
		},
		{
			name:   "Zero ID",
			userID: 0,
			want:   false,
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.name, func(t *testing.T) {
			db := newTestDB(t)
			m := UserModel{db}

			exists, err := m.Exists(subtest.userID)

			assert.Equal(t, exists, subtest.want)
			assert.NilError(t, err)
		})
	}
}
