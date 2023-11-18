package main

import (
	"testing"
	"time"

	"github.com/vladComan0/letsgo/internal/assert"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name     string
		tm       time.Time
		expected string
	}{
		{
			name:     "UTC",
			tm:       time.Date(2022, 3, 17, 10, 15, 0, 0, time.UTC),
			expected: "17 Mar 2022 at 10:15",
		},
		{
			name:     "Empty",
			tm:       time.Time{},
			expected: "",
		},
		{
			name:     "CET",
			tm:       time.Date(2022, 3, 17, 16, 22, 0, 0, time.FixedZone("CET", 1*60*60)),
			expected: "17 Mar 2022 at 15:22",
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.name, func(t *testing.T) {
			actual := humanDate(subtest.tm)
			assert.Equal(t, actual, subtest.expected)
		})
	}
}
