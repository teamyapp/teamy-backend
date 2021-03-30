package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	num1 := 3
	num2 := 5
	sum := Add(3, 5)

	assert.Equal(t, 8, sum, "Sum was incorrect, got: %d, want: %d.", num1, num2)
}
