package app

import "testing"

func TestAdd(t *testing.T) {
	num1 := RandInt()
	num2 := RandInt()
	total := Add(num1, num2)
	if total != num1 + num2 {
		t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
	}
}
